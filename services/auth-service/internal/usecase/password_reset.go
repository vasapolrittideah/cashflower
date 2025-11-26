package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/config"
	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/model"
	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/repository"
	authtypes "github.com/vasapolrittideah/money-tracker-api/services/auth-service/pkg/types"
	"github.com/vasapolrittideah/money-tracker-api/shared/auth"
	"github.com/vasapolrittideah/money-tracker-api/shared/mailer"
	"github.com/vasapolrittideah/money-tracker-api/shared/security"
)

// PasswordResetUsecase defines the business logic for password reset token operations.
type PasswordResetUsecase interface {
	// RequestPasswordReset initiates the password reset process for a given email.
	RequestPasswordReset(ctx context.Context, email string) error

	// ResetPassword resets the user's password using the provided jti and new password.
	ResetPassword(ctx context.Context, jti, newPassword string) error

	// ValidatePasswordResetToken checks if the provided jti is not used.
	ValidatePasswordResetToken(ctx context.Context, jti string) error
}

type passwordResetUsecase struct {
	userRepo       repository.UserRepository
	tokenRepo      repository.PasswordResetTokenRepository
	jwtAuth        auth.JWTAuthenticator
	mailer         *mailer.Mailer
	authServiceCfg *config.AuthServiceConfig
}

var (
	ErrTokenNotFound    = errors.New("password reset token not found")
	ErrTokenAlreadyUsed = errors.New("password reset token has already been used")
	ErrTokenExpired     = errors.New("password reset token has expired")
	ErrInvalidToken     = errors.New("invalid password reset token")
)

// NewPasswordResetUsecase creates a new instance of PasswordResetUsecase.
func NewPasswordResetUsecase(
	userRepo repository.UserRepository,
	tokenRepo repository.PasswordResetTokenRepository,
	jwtAuth auth.JWTAuthenticator,
	mailer *mailer.Mailer,
	authServiceCfg *config.AuthServiceConfig,
) PasswordResetUsecase {
	return &passwordResetUsecase{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		jwtAuth:        jwtAuth,
		mailer:         mailer,
		authServiceCfg: authServiceCfg,
	}
}

func (u *passwordResetUsecase) RequestPasswordReset(ctx context.Context, email string) error {
	// Get user by email
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// To prevent email enumeration, do not reveal that the email does not exist.
			return nil
		}
		return err
	}

	// Invalidate any existing unused tokens for this user
	if err := u.tokenRepo.InvalidateUserTokens(ctx, user.ID.Hex()); err != nil {
		return err
	}

	// Generate password reset token with JTI
	tokenStr, jti, err := u.generatePasswordResetToken(user.ID.Hex(), user.Email)
	if err != nil {
		return err
	}

	// Store token in database
	resetToken := &model.PasswordResetToken{
		JTI:       jti,
		UserID:    user.ID,
		Email:     user.Email,
		Used:      false,
		ExpiresAt: time.Now().Add(u.authServiceCfg.Token.PasswordResetTokenExpiresIn),
	}

	if _, err := u.tokenRepo.CreateToken(ctx, resetToken); err != nil {
		return err
	}

	// Send email with the reset link
	resetLink := fmt.Sprintf("%s?token=%s", u.authServiceCfg.AppPasswordResetURL, tokenStr)
	htmlBody := fmt.Sprintf(`
		<p>Hi,</p>
		<p>We received a request to reset the password for your account.</p>
		<p>If you made this request, please click the link below to create a new password:</p>

		<p><a href="%s">%s</a></p>

		<p>This link will expire in %s for your security.</p>
		<p>If you did not request a password reset, you can safely ignore this email—your account will remain secure.</p>

		<p>Thank you,</p>
		<p>Money Tracker Team</p>
	`, resetLink, resetLink, u.authServiceCfg.Token.PasswordResetTokenExpiresIn)

	if err := u.mailer.SendHTML([]string{user.Email}, "Password Reset Request", htmlBody); err != nil {
		return err
	}

	return nil
}

func (u *passwordResetUsecase) ResetPassword(ctx context.Context, jti, newPassword string) error {
	// Check token in database
	resetToken, err := u.tokenRepo.GetTokenByJTI(ctx, jti)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrTokenNotFound
		}
		return err
	}

	// Validate token status
	if resetToken.Used {
		return ErrTokenAlreadyUsed
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return ErrTokenExpired
	}

	// Hash new password
	passwordHash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user's password
	if _, err := u.userRepo.UpdateUser(ctx, resetToken.UserID.Hex(), repository.UpdateUserParams{
		PasswordHash: &passwordHash,
	}); err != nil {
		return err
	}

	// Mark token as used
	if err := u.tokenRepo.MarkTokenAsUsed(ctx, jti); err != nil {
		return err
	}

	return nil
}

func (u *passwordResetUsecase) ValidatePasswordResetToken(ctx context.Context, jti string) error {
	// Check token in database
	resetToken, err := u.tokenRepo.GetTokenByJTI(ctx, jti)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrTokenNotFound
		}
		return err
	}

	// Validate token status
	if resetToken.Used {
		return ErrTokenAlreadyUsed
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return ErrTokenExpired
	}

	return nil
}

// generatePasswordResetToken creates a password reset JWT token with a unique JTI.
func (u *passwordResetUsecase) generatePasswordResetToken(userID, email string) (string, string, error) {
	jti, err := generateJTI()
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	claims := authtypes.PasswordResetClaims{
		UserID: userID,
		Email:  email,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    u.authServiceCfg.Token.Issuer,
			Audience:  jwt.ClaimStrings{u.authServiceCfg.Token.Issuer},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(u.authServiceCfg.Token.PasswordResetTokenExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	tokenStr, err := u.jwtAuth.GenerateToken(claims, u.authServiceCfg.Token.PasswordResetTokenSecret)
	if err != nil {
		return "", "", err
	}

	return tokenStr, jti, nil
}

// generateJTI generates a unique JTI.
func generateJTI() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
