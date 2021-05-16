package models

import uuid "github.com/gofrs/uuid"

type GetSettingsModel struct {
	Type    string      `json:"type"`
	UserIds []uuid.UUID `json:"userIds"`
}
