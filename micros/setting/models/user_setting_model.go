package models

import (
	uuid "github.com/gofrs/uuid"
)

type UserSettingModel struct {
	ObjectId    uuid.UUID `json:"objectId"`
	OwnerUserId uuid.UUID `json:"ownerUserId"`
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	Type        string    `json:"type"`
	IsSystem    bool      `json:"isSystem"`
	CreatedDate int64     `json:"created_date"`
}
