package handler

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/usecase"
	"github.com/vasapolrittideah/money-tracker-api/shared/interceptor"
	authpbv1 "github.com/vasapolrittideah/money-tracker-api/shared/protos/auth/v1"
)

func (h *authGRPCHandler) RequestPasswordReset(
	ctx context.Context,
	req *authpbv1.RequestPasswordResetRequest,
) (*authpbv1.RequestPasswordResetResponse, error) {
	email := req.GetEmail()
	if email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}

	err := h.passwordResetUsecase.RequestPasswordReset(ctx, email)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to request password reset")
		return nil, status.Errorf(codes.Internal, "something went wrong")
	}

	return &authpbv1.RequestPasswordResetResponse{}, nil
}

func (h *authGRPCHandler) ResetPassword(
	ctx context.Context,
	req *authpbv1.ResetPasswordRequest,
) (*authpbv1.ResetPasswordResponse, error) {
	newPassword := req.GetNewPassword()
	if newPassword == "" {
		return nil, status.Errorf(codes.InvalidArgument, "new password is required")
	}

	claims, ok := ctx.Value(interceptor.UserClaimsKey).(jwt.MapClaims)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid password reset token claims")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid JTI claim")
	}

	err := h.passwordResetUsecase.ResetPassword(ctx, jti, newPassword)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to reset password")

		switch {
		case errors.Is(err, usecase.ErrTokenNotFound):
			return nil, status.Errorf(codes.NotFound, "password reset token not found")
		case errors.Is(err, usecase.ErrTokenAlreadyUsed):
			return nil, status.Errorf(codes.FailedPrecondition, "password reset token has already been used")
		case errors.Is(err, usecase.ErrTokenExpired):
			return nil, status.Errorf(codes.Unauthenticated, "password reset token has expired")
		case errors.Is(err, usecase.ErrInvalidToken):
			return nil, status.Errorf(codes.Unauthenticated, "invalid password reset token")
		default:
			return nil, status.Errorf(codes.Internal, "something went wrong")
		}
	}

	return &authpbv1.ResetPasswordResponse{}, nil
}

func (h *authGRPCHandler) ValidatePasswordResetToken(
	ctx context.Context,
	_ *authpbv1.ValidatePasswordResetTokenRequest,
) (*authpbv1.ValidatePasswordResetTokenResponse, error) {
	claims, ok := ctx.Value(interceptor.UserClaimsKey).(jwt.MapClaims)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid password reset token claims")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid JTI claim")
	}

	err := h.passwordResetUsecase.ValidatePasswordResetToken(ctx, jti)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to validate password reset token")

		switch {
		case errors.Is(err, usecase.ErrTokenNotFound):
			return nil, status.Errorf(codes.NotFound, "password reset token not found")
		case errors.Is(err, usecase.ErrTokenAlreadyUsed):
			return nil, status.Errorf(codes.FailedPrecondition, "password reset token has already been used")
		case errors.Is(err, usecase.ErrTokenExpired):
			return nil, status.Errorf(codes.Unauthenticated, "password reset token has expired")
		case errors.Is(err, usecase.ErrInvalidToken):
			return nil, status.Errorf(codes.Unauthenticated, "invalid password reset token")
		default:
			return nil, status.Errorf(codes.Internal, "something went wrong")
		}
	}

	return &authpbv1.ValidatePasswordResetTokenResponse{}, nil
}
