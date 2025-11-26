package handler

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/usecase"
	authpbv1 "github.com/vasapolrittideah/money-tracker-api/shared/protos/auth/v1"
)

type authGRPCHandler struct {
	authpbv1.UnimplementedAuthServiceServer

	logger               *zerolog.Logger
	authUsecase          usecase.AuthUsecase
	passwordResetUsecase usecase.PasswordResetUsecase
}

func NewAuthGRPCHandler(
	server *grpc.Server,
	logger *zerolog.Logger,
	authUsecase usecase.AuthUsecase,
	passwordResetUsecase usecase.PasswordResetUsecase,
) authpbv1.AuthServiceServer {
	handler := &authGRPCHandler{
		logger:               logger,
		authUsecase:          authUsecase,
		passwordResetUsecase: passwordResetUsecase,
	}
	authpbv1.RegisterAuthServiceServer(server, handler)

	return handler
}

func (h *authGRPCHandler) Login(ctx context.Context, req *authpbv1.LoginRequest) (*authpbv1.LoginResponse, error) {
	params := usecase.LoginParams{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	tokens, err := h.authUsecase.Login(ctx, params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to login")

		switch {
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
		default:
			return nil, status.Errorf(codes.Internal, "something went wrong")
		}
	}

	return &authpbv1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *authGRPCHandler) Register(
	ctx context.Context,
	req *authpbv1.RegisterRequest,
) (*authpbv1.RegisterResponse, error) {
	params := usecase.RegisterParams{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	tokens, err := h.authUsecase.Register(ctx, params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to register")

		switch {
		case errors.Is(err, usecase.ErrUserAlreadyExists):
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		default:
			return nil, status.Errorf(codes.Internal, "something went wrong")
		}
	}

	return &authpbv1.RegisterResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
