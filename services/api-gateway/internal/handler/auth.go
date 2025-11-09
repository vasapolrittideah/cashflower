package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/vasapolrittideah/money-tracker-api/services/api-gateway/internal/payload"
	authclient "github.com/vasapolrittideah/money-tracker-api/services/auth-service/pkg/client"
	authpbv1 "github.com/vasapolrittideah/money-tracker-api/shared/protos/auth/v1"
	"github.com/vasapolrittideah/money-tracker-api/shared/utilities"
	"github.com/vasapolrittideah/money-tracker-api/shared/validator"
)

type AuthHTTPHandler struct {
	router            *chi.Mux
	logger            *zerolog.Logger
	authServiceClient *authclient.AuthServiceClient
}

func NewAuthHTTPHandler(
	router *chi.Mux,
	logger *zerolog.Logger,
	authServiceClient *authclient.AuthServiceClient,
) *AuthHTTPHandler {
	handler := &AuthHTTPHandler{
		router:            router,
		logger:            logger,
		authServiceClient: authServiceClient,
	}

	return handler
}

func (h *AuthHTTPHandler) RegisterRoutes() {
	h.router.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.login)
		r.Post("/register", h.register)
	})
}

func (h *AuthHTTPHandler) login(w http.ResponseWriter, r *http.Request) {
	var req payload.LoginRequest
	if err := utilities.ReadJSON(w, r, &req); err != nil {
		utilities.WriteRequestErrorResponse(w, r, err.Error(), h.logger)
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		utilities.WriteValidationErrorResponse(w, r, errs, h.logger)
		return
	}

	grpcResp, err := h.authServiceClient.Client.Login(r.Context(), &authpbv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		utilities.WriteInternalErrorResponse(w, r, err, h.logger)
		return
	}

	payload := &payload.LoginResponse{
		AccessToken:  grpcResp.AccessToken,
		RefreshToken: grpcResp.RefreshToken,
	}

	utilities.WriteSuccessResponse(w, r, payload, h.logger)
}

func (h *AuthHTTPHandler) register(w http.ResponseWriter, r *http.Request) {
	var req payload.RegisterRequest
	if err := utilities.ReadJSON(w, r, &req); err != nil {
		utilities.WriteRequestErrorResponse(w, r, err.Error(), h.logger)
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		utilities.WriteValidationErrorResponse(w, r, errs, h.logger)
		return
	}

	grpcResp, err := h.authServiceClient.Client.Register(r.Context(), &authpbv1.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		utilities.WriteInternalErrorResponse(w, r, err, h.logger)
		return
	}

	payload := &payload.RegisterResponse{
		AccessToken:  grpcResp.AccessToken,
		RefreshToken: grpcResp.RefreshToken,
	}

	utilities.WriteSuccessResponse(w, r, payload, h.logger)
}
