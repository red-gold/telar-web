package models

import uuid "github.com/gofrs/uuid"

type UpdateSettingGroupItemModel struct {
	ObjectId uuid.UUID `json:"objectId"`
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	IsSystem bool      `json:"isSystem"`
}
