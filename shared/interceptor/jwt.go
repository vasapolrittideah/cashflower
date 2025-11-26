package interceptor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/vasapolrittideah/money-tracker-api/shared/auth"
)

type contextKey struct{}

var UserClaimsKey = contextKey{}

func NewJWTInterceptor(
	jwtAuth auth.JWTAuthenticator,
	secret string,
	exemptMethods []string,
) grpc.UnaryServerInterceptor {
	exemptMap := make(map[string]bool)
	for _, method := range exemptMethods {
		exemptMap[method] = true
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Skip authentication for exempt methods
		if exemptMap[info.FullMethod] {
			return handler(ctx, req)
		}

		claims, err := extractAndValidateJWT(ctx, jwtAuth, secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		ctx = context.WithValue(ctx, UserClaimsKey, claims)

		return handler(ctx, req)
	}
}

func extractAndValidateJWT(ctx context.Context, jwtAuth auth.JWTAuthenticator, secret string) (jwt.MapClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}

	fmt.Println(md)

	authHeaders := md.Get("Authorization")
	if len(authHeaders) == 0 {
		return nil, errors.New("missing authorization header")
	}

	authHeader := authHeaders[0]
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, errors.New("invalid authorization header format")
	}

	tokenString := parts[1]

	claims := jwt.MapClaims{}
	_, err := jwtAuth.ValidateTokenWithClaims(tokenString, secret, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
