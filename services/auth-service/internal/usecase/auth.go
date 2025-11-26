package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/config"
	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/model"
	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/repository"
	authtypes "github.com/vasapolrittideah/money-tracker-api/services/auth-service/pkg/types"
	"github.com/vasapolrittideah/money-tracker-api/shared/auth"
	"github.com/vasapolrittideah/money-tracker-api/shared/security"
)

// AuthUsecase defines the interface for authentication-related use cases.
type AuthUsecase interface {
	Login(ctx context.Context, params LoginParams) (*authtypes.Tokens, error)
	Register(ctx context.Context, params RegisterParams) (*authtypes.Tokens, error)
}

// LoginParams defines the parameters for user login.
type LoginParams struct {
	Email    string
	Password string
}

// RegisterParams defines the parameters for user registration.
type RegisterParams struct {
	Email    string
	Password string
}

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type authUsecase struct {
	identityRepo   repository.IdentityRepository
	sessionRepo    repository.SessionRepository
	userRepo       repository.UserRepository
	jwtAuth        auth.JWTAuthenticator
	authServiceCfg *config.AuthServiceConfig
}

func NewAuthUsecase(
	identityRepo repository.IdentityRepository,
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
	jwtAuth auth.JWTAuthenticator,
	authServiceCfg *config.AuthServiceConfig,
) AuthUsecase {
	return &authUsecase{
		identityRepo:   identityRepo,
		sessionRepo:    sessionRepo,
		userRepo:       userRepo,
		jwtAuth:        jwtAuth,
		authServiceCfg: authServiceCfg,
	}
}

func (u *authUsecase) Login(ctx context.Context, params LoginParams) (*authtypes.Tokens, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	if ok, err := security.VerifyPassword(params.Password, user.PasswordHash); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrInvalidCredentials
	}

	if err := u.identityRepo.UpdateLastLogin(ctx, user.ID.Hex()); err != nil {
		return nil, err
	}

	return u.createAuthSession(ctx, user.ID.Hex())
}

func (u *authUsecase) Register(ctx context.Context, params RegisterParams) (*authtypes.Tokens, error) {
	passwordHash, err := security.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.CreateUser(ctx, &model.User{
		Email:        params.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}

		return nil, err
	}

	if _, err := u.identityRepo.CreateIdentity(ctx, &model.Identity{
		UserID:     user.ID.Hex(),
		Provider:   "email",
		ProviderID: "",
		Email:      user.Email,
	}); err != nil {
		return nil, err
	}

	return u.createAuthSession(ctx, user.ID.Hex())
}

func (u *authUsecase) createAuthSession(ctx context.Context, userID string) (*authtypes.Tokens, error) {
	session, err := u.sessionRepo.CreateSession(ctx, &model.Session{UserID: userID})
	if err != nil {
		return nil, err
	}

	accessToken, err := u.generateToken(
		userID,
		session.ID.Hex(),
		u.authServiceCfg.Token.AccessTokenSecret,
		u.authServiceCfg.Token.AccessTokenExpiresIn,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := u.generateToken(
		userID,
		session.ID.Hex(),
		u.authServiceCfg.Token.RefreshTokenSecret,
		u.authServiceCfg.Token.RefreshTokenExpiresIn,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if _, err := u.sessionRepo.UpdateTokens(ctx, session.ID.Hex(), repository.UpdateTokensParams{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  now.Add(u.authServiceCfg.Token.AccessTokenExpiresIn),
		RefreshTokenExpiresAt: now.Add(u.authServiceCfg.Token.RefreshTokenExpiresIn),
	}); err != nil {
		return nil, err
	}

	return &authtypes.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *authUsecase) generateToken(userID, sessionID, secret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := authtypes.JWTClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    u.authServiceCfg.Token.Issuer,
			Audience:  jwt.ClaimStrings{u.authServiceCfg.Token.Issuer},
		},
	}
	token, err := u.jwtAuth.GenerateToken(claims, secret)
	if err != nil {
		return "", err
	}

	return token, nil
}
