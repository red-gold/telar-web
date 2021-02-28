package models

import uuid "github.com/gofrs/uuid"

type UserRegisterModel struct {
	ObjectId        uuid.UUID `json:"objectId"`
	Username        string    `bson:"username" json:"username"`
	Password        string    `json:"password"`
	ConfirmPassword string    `json:"confirmPassword"`
}
