package dto

import uuid "github.com/gofrs/uuid"

type UserSetting struct {
	ObjectId    uuid.UUID `json:"objectId" bson:"objectId"`
	CreatedDate int64     `json:"created_date" bson:"created_date"`
	OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	Name        string    `json:"name" bson:"name"`
	Value       string    `json:"value" bson:"value"`
	Type        string    `json:"type" bson:"type"`
	IsSystem    bool      `json:"isSystem" bson:"isSystem"`
}
