package models

import (
	uuid "github.com/gofrs/uuid"
)

type UpdateLastSeenModel struct {
	UserId uuid.UUID `json:"userId" bson:"userId"`
}
