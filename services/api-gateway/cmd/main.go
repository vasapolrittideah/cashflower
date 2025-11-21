package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/vasapolrittideah/money-tracker-api/services/api-gateway/internal/config"
	"github.com/vasapolrittideah/money-tracker-api/services/api-gateway/internal/handler"
	authclient "github.com/vasapolrittideah/money-tracker-api/services/auth-service/pkg/client"
	"github.com/vasapolrittideah/money-tracker-api/shared/discovery"
	"github.com/vasapolrittideah/money-tracker-api/shared/logger"
)

func main() {
	logger := logger.New()

	apiGatewayCfg := config.NewAPIGatewayConfig(logger)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	server := &http.Server{
		Addr:         apiGatewayCfg.Address,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		IdleTimeout:  time.Minute,
		Handler:      r,
	}

	consulRegistry, err := discovery.NewConsulRegistry(logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create Consul registry")
	}

	authServiceClient, err := authclient.NewAuthServiceClient(
		"auth-service",
		consulRegistry,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create auth service client")
	}

	authHandler := handler.NewAuthHTTPHandler(logger, authServiceClient)
	r.Route("/api/v1", func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info().Str("address", apiGatewayCfg.Address).Msg("starting HTTP server...")
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	select {
	case err := <-serverErrors:
		logger.Error().Err(err).Msg("failed to start HTTP server")

	case sig := <-shutdown:
		logger.Info().Interface("signal", sig).Msg("shutting down HTTP server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("failed to shutdown HTTP server")
		}
	}
}
