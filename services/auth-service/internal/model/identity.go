package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Identity represents a user's identity in the authentication system.
// It stores the mapping between a user and their identities from both external
// providers (Google, Facebook, etc.) and local authentication (email and password).
type Identity struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	UserID      string        `bson:"user_id"`
	ProviderID  string        `bson:"provider_id"`
	Provider    string        `bson:"provider"`
	Email       string        `bson:"email"`
	LastLoginAt time.Time     `bson:"last_login_at"`
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
}
