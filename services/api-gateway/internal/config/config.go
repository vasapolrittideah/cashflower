package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
)

type APIGatewayConfig struct {
	Environment string `env:"ENVIRONMENT"`
	Name        string `env:"SERVICE_NAME"`
	Address     string `env:"SERVICE_ADDRESS"`
}

func NewAPIGatewayConfig(logger *zerolog.Logger) *APIGatewayConfig {
	cfg, err := env.ParseAs[APIGatewayConfig]()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse environment variables")
	}

	return &cfg
}
