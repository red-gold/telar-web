package models

import "github.com/gofrs/uuid"

type DispatchProfilesModel struct {
	UserIds   []uuid.UUID `json:"userIds" bson:"userIds"`
	ReqUserId uuid.UUID   `json:"reqUserId" bson:"reqUserId"`
}
