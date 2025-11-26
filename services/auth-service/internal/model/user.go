package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User represents a user in the authentication system.
type User struct {
	ID                        bson.ObjectID `bson:"_id,omitempty"`
	Email                     string        `bson:"email"`
	PasswordHash              string        `bson:"password_hash"`
	Verified                  bool          `bson:"verified"`
	VerificationCode          string        `bson:"verification_code"`
	VerificationCodeExpiresAt time.Time     `bson:"verification_code_expires_at"`
	CreatedAt                 time.Time     `bson:"created_at"`
	UpdatedAt                 time.Time     `bson:"updated_at"`
}
