package models

import (
	"time"

	"github.com/google/uuid"
)

type UserAuthQueryModel struct {
	UserUID      uuid.UUID `json:"uid"`
	Username     string    `json:"username"`
	Password     []byte    `json:"password"`
	AccessToken  string    `json:"access_token"`
	Role         string    `json:"role"`
	TokenExpires int       `json:"token_expires"`
	CreatedDate  time.Time `json:"created_date"`
	LastUpdated  time.Time `json:"last_updated"`
}
