package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/config"
	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/domain"
	authtypes "github.com/vasapolrittideah/money-tracker-api/services/auth-service/pkg/types"
	"github.com/vasapolrittideah/money-tracker-api/shared/auth"
	"github.com/vasapolrittideah/money-tracker-api/shared/security"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type authUsecase struct {
	identityRepo   domain.IdentityRepository
	sessionRepo    domain.SessionRepository
	userRepo       domain.UserRepository
	authenticator  auth.Authenticator
	authServiceCfg *config.AuthServiceConfig
}

func NewAuthUsecase(
	identityRepo domain.IdentityRepository,
	sessionRepo domain.SessionRepository,
	userRepo domain.UserRepository,
	authenticator auth.Authenticator,
	authServiceCfg *config.AuthServiceConfig,
) domain.AuthUsecase {
	return &authUsecase{
		identityRepo:   identityRepo,
		sessionRepo:    sessionRepo,
		userRepo:       userRepo,
		authenticator:  authenticator,
		authServiceCfg: authServiceCfg,
	}
}

func (u *authUsecase) Login(ctx context.Context, params domain.LoginParams) (*authtypes.Tokens, error) {
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

func (u *authUsecase) Register(ctx context.Context, params domain.RegisterParams) (*authtypes.Tokens, error) {
	passwordHash, err := security.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.CreateUser(ctx, &domain.User{
		Email:        params.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}

		return nil, err
	}

	if _, err := u.identityRepo.CreateIdentity(ctx, &domain.Identity{
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
	session, err := u.sessionRepo.CreateSession(ctx, &domain.Session{UserID: userID})
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
	if _, err := u.sessionRepo.UpdateTokens(ctx, session.ID.Hex(), domain.UpdateTokensParams{
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
	token, err := u.authenticator.GenerateToken(claims, secret)
	if err != nil {
		return "", err
	}

	return token, nil
}
