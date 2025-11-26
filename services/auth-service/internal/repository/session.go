package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/model"
)

// SessionRepository defines the interface for session-related database operations.
type SessionRepository interface {
	CreateSession(ctx context.Context, session *model.Session) (*model.Session, error)
	GetSessionByUserID(ctx context.Context, userID string) (*model.Session, error)
	UpdateTokens(ctx context.Context, id string, params UpdateTokensParams) (*model.Session, error)
}

// UpdateTokensParams defines the parameters for updating session tokens.
type UpdateTokensParams struct {
	AccessToken           string    `bson:"access_token"`
	RefreshToken          string    `bson:"refresh_token"`
	AccessTokenExpiresAt  time.Time `bson:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `bson:"refresh_token_expires_at"`
}

const sessionCollection = "sessions"

type sessionMongoRepository struct {
	db *mongo.Database
}

func NewSessionMongoRepository(db *mongo.Database) SessionRepository {
	return &sessionMongoRepository{db: db}
}

func (r *sessionMongoRepository) CreateSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	result, err := r.db.Collection(sessionCollection).InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}

	if objectID, ok := result.InsertedID.(bson.ObjectID); ok {
		session.ID = objectID
	} else {
		return nil, errors.New("failed to convert inserted ID to ObjectID")
	}

	return session, nil
}

func (r *sessionMongoRepository) GetSessionByUserID(ctx context.Context, userID string) (*model.Session, error) {
	result := r.db.Collection(sessionCollection).FindOne(ctx, bson.M{"user_id": userID})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var session model.Session
	if err := result.Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *sessionMongoRepository) UpdateTokens(
	ctx context.Context,
	id string,
	params UpdateTokensParams,
) (*model.Session, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result := r.db.Collection(sessionCollection).FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": params},
	)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var session model.Session
	if err := result.Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}
