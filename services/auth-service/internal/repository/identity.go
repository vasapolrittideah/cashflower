package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/model"
)

// IdentityRepository defines the interface for identity-related database operations.
type IdentityRepository interface {
	CreateIdentity(ctx context.Context, identity *model.Identity) (*model.Identity, error)
	GetIdentitiesByUserID(ctx context.Context, userID string) ([]model.Identity, error)
	GetIdentityByProvider(ctx context.Context, providerID string, provider string) (*model.Identity, error)
	UpdateLastLogin(ctx context.Context, userID string) error
}

const identityCollection = "identities"

type identityMongoRepository struct {
	db *mongo.Database
}

func NewIdentityMongoRepository(db *mongo.Database) IdentityRepository {
	return &identityMongoRepository{db: db}
}

func (r *identityMongoRepository) CreateIdentity(
	ctx context.Context,
	identity *model.Identity,
) (*model.Identity, error) {
	now := time.Now()
	identity.CreatedAt = now
	identity.UpdatedAt = now

	result, err := r.db.Collection(identityCollection).InsertOne(ctx, identity)
	if err != nil {
		return nil, err
	}

	if objectID, ok := result.InsertedID.(bson.ObjectID); ok {
		identity.ID = objectID
	} else {
		return nil, errors.New("failed to convert inserted ID to ObjectID")
	}

	return identity, nil
}

func (r *identityMongoRepository) GetIdentitiesByUserID(ctx context.Context, userID string) ([]model.Identity, error) {
	cursor, err := r.db.Collection(identityCollection).Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	var identities []model.Identity
	if err := cursor.All(ctx, &identities); err != nil {
		return nil, err
	}

	return identities, nil
}

func (r *identityMongoRepository) GetIdentityByProvider(
	ctx context.Context,
	providerID string,
	provider string,
) (*model.Identity, error) {
	result := r.db.Collection(identityCollection).FindOne(ctx, bson.M{
		"provider_id": providerID,
		"provider":    provider,
	})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var identity model.Identity
	if err := result.Decode(&identity); err != nil {
		return nil, err
	}

	return &identity, nil
}

func (r *identityMongoRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.db.Collection(identityCollection).UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"last_login_at": time.Now()}},
	)
	return err
}
