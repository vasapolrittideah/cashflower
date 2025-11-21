package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/domain"
)

const sessionCollection = "sessions"

type sessionMongoRepository struct {
	db *mongo.Database
}

func NewSessionMongoRepository(db *mongo.Database) domain.SessionRepository {
	return &sessionMongoRepository{db: db}
}

func (r *sessionMongoRepository) CreateSession(ctx context.Context, session *domain.Session) (*domain.Session, error) {
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

func (r *sessionMongoRepository) GetSessionByUserID(ctx context.Context, userID string) (*domain.Session, error) {
	result := r.db.Collection(sessionCollection).FindOne(ctx, bson.M{"user_id": userID})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var session domain.Session
	if err := result.Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *sessionMongoRepository) UpdateTokens(
	ctx context.Context,
	id string,
	params domain.UpdateTokensParams,
) (*domain.Session, error) {
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

	var session domain.Session
	if err := result.Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}
