package repository

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/model"
)

// PasswordResetRepository defines the interface for password reset token operations.
type PasswordResetTokenRepository interface {
	// CreateToken creates a new password reset token.
	CreateToken(ctx context.Context, token *model.PasswordResetToken) (*model.PasswordResetToken, error)

	// GetTokenByJTI retrieves a token by its JTI.
	GetTokenByJTI(ctx context.Context, jti string) (*model.PasswordResetToken, error)

	// MarkTokenAsUsed marks a token as used.
	MarkTokenAsUsed(ctx context.Context, jti string) error

	// DeleteExpiredTokens removes expired tokens from the database.
	DeleteExpiredTokens(ctx context.Context) (int64, error)

	// InvalidateUserTokens invalidates all unused tokens for a specific user.
	InvalidateUserTokens(ctx context.Context, userID string) error
}

const passwordResetTokenCollection = "password_reset_tokens"

type passwordResetTokenMongoRepository struct {
	db *mongo.Database
}

// NewPasswordResetTokenMongoRepository creates a new MongoDB repository for password reset tokens.
func NewPasswordResetTokenMongoRepository(
	ctx context.Context,
	logger *zerolog.Logger,
	db *mongo.Database,
) PasswordResetTokenRepository {
	collection := db.Collection(passwordResetTokenCollection)

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "jti", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "email", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0), // TTL index
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create password reset token indexes")
	}

	return &passwordResetTokenMongoRepository{
		db: db,
	}
}

func (r *passwordResetTokenMongoRepository) CreateToken(
	ctx context.Context,
	token *model.PasswordResetToken,
) (*model.PasswordResetToken, error) {
	now := time.Now()
	token.CreatedAt = now
	token.UpdatedAt = now
	token.Used = false

	result, err := r.db.Collection(passwordResetTokenCollection).InsertOne(ctx, token)
	if err != nil {
		return nil, err
	}

	if objectID, ok := result.InsertedID.(bson.ObjectID); ok {
		token.ID = objectID
	}

	return token, nil
}

func (r *passwordResetTokenMongoRepository) GetTokenByJTI(
	ctx context.Context,
	jti string,
) (*model.PasswordResetToken, error) {
	filter := bson.M{"jti": jti}

	var token model.PasswordResetToken
	err := r.db.Collection(passwordResetTokenCollection).FindOne(ctx, filter).Decode(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *passwordResetTokenMongoRepository) MarkTokenAsUsed(ctx context.Context, jti string) error {
	filter := bson.M{"jti": jti}
	update := bson.M{
		"$set": bson.M{
			"used":       true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.db.Collection(passwordResetTokenCollection).UpdateOne(ctx, filter, update)
	return err
}

func (r *passwordResetTokenMongoRepository) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	}

	result, err := r.db.Collection(passwordResetTokenCollection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (r *passwordResetTokenMongoRepository) InvalidateUserTokens(ctx context.Context, userID string) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"user_id": objectID,
		"used":    false,
	}
	update := bson.M{
		"$set": bson.M{
			"used":       true,
			"updated_at": time.Now(),
		},
	}

	_, err = r.db.Collection(passwordResetTokenCollection).UpdateMany(ctx, filter, update)
	return err
}
