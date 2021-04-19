package models

import "github.com/gofrs/uuid"

type GetProfilesModel struct {
	UserIds []uuid.UUID `json:"userIds" bson:"userIds"`
}
