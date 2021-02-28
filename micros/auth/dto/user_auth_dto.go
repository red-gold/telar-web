package dto

import (
	uuid "github.com/gofrs/uuid"
)

type UserAuth struct {
	ObjectId      uuid.UUID `json:"objectId" bson:"objectId"`
	Username      string    `json:"username" bson:"username"`
	Password      []byte    `json:"password" bson:"password"`
	AccessToken   string    `json:"access_token" bson:"access_token"`
	EmailVerified bool      `json:"emailVerified" bson:"emailVerified"`
	Role          string    `json:"role" bson:"role"`
	PhoneVerified bool      `json:"phoneVerified" bson:"phoneVerified"`
	TokenExpires  int64     `json:"token_expires" bson:"token_expires"`
	CreatedDate   int64     `json:"created_date" bson:"created_date"`
	LastUpdated   int64     `json:"last_updated" bson:"last_updated"`
}
