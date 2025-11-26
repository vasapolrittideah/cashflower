package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// PasswordResetToken represents a password reset token with JTI (JWT Token Identifier).
type PasswordResetToken struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"user_id"`
	JTI       string        `bson:"jti"`
	Email     string        `bson:"email"`
	Used      bool          `bson:"used"`
	ExpiresAt time.Time     `bson:"expires_at"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}
