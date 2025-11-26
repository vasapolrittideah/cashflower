package repository

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/vasapolrittideah/money-tracker-api/services/auth-service/internal/model"
)

// UserRepository defines the interface for user-related database operations.
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, params UpdateUserParams) (*model.User, error)
	DeleteUser(ctx context.Context, id string) (*model.User, error)
	ListUsers(ctx context.Context, params FilterUsersParams) ([]*model.User, error)
}

// UpdateUserParams defines the optional parameters for updating a user.
// Only the fields that are not nil will be updated.
type UpdateUserParams struct {
	Email        *string
	PasswordHash *string
}

// FilterUsersParams defines the parameters for filtering and paginating users.
type FilterUsersParams struct {
	Email    *string
	Verified *bool
	Limit    uint64
	Offset   uint64
	SortBy   *string
	SortDesc bool
}

const userCollection = "users"

type userMongoRepository struct {
	db *mongo.Database
}

func NewUserMongoRepository(ctx context.Context, logger *zerolog.Logger, db *mongo.Database) UserRepository {
	collection := db.Collection(userCollection)

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create user indexes")
	}

	return &userMongoRepository{db: db}
}

func (r *userMongoRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := r.db.Collection(userCollection).InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	if objectID, ok := result.InsertedID.(bson.ObjectID); ok {
		user.ID = objectID
	} else {
		return nil, errors.New("failed to convert inserted ID to ObjectID")
	}

	return user, nil
}

func (r *userMongoRepository) GetUser(ctx context.Context, id string) (*model.User, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result := r.db.Collection(userCollection).FindOne(ctx, bson.M{"_id": objectID})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var user model.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userMongoRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	result := r.db.Collection(userCollection).FindOne(ctx, bson.M{"email": email})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var user model.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userMongoRepository) UpdateUser(
	ctx context.Context,
	id string,
	params UpdateUserParams,
) (*model.User, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Build update query
	updateMap := bson.M{}
	if params.Email != nil {
		updateMap["email"] = params.Email
	}
	if params.PasswordHash != nil {
		updateMap["password_hash"] = params.PasswordHash
	}

	if len(updateMap) == 0 {
		return nil, errors.New("no user fields to update")
	}

	updateMap["updated_at"] = time.Now()

	result := r.db.Collection(userCollection).FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updateMap},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var user model.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userMongoRepository) DeleteUser(ctx context.Context, id string) (*model.User, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result := r.db.Collection(userCollection).FindOneAndDelete(ctx, bson.M{"_id": objectID})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var user model.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userMongoRepository) ListUsers(ctx context.Context, params FilterUsersParams) ([]*model.User, error) {
	findOptions := options.Find()

	limit := params.Limit
	if limit == 0 {
		limit = 10
	}
	findOptions.SetLimit(int64(limit))

	if params.Offset > 0 {
		findOptions.SetSkip(int64(params.Offset))
	}

	sortBy := "created_at"
	if params.SortBy != nil {
		sortBy = *params.SortBy
	}

	sortOrder := -1
	if !params.SortDesc {
		sortOrder = 1
	}
	findOptions.SetSort(bson.D{{Key: sortBy, Value: sortOrder}})

	// Build filter query
	filter := bson.M{}
	if params.Email != nil {
		filter["email"] = *params.Email
	}
	if params.Verified != nil {
		filter["verified"] = *params.Verified
	}

	cursor, err := r.db.Collection(userCollection).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
