package dto

import (
	uuid "github.com/gofrs/uuid"
)

type ActionRoom struct {
	ObjectId    uuid.UUID `json:"objectId" bson:"objectId"`
	OwnerUserId uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	PrivateKey  string    `json:"privateKey" bson:"privateKey"`
	AccessKey   string    `json:"accessKey" bson:"accessKey"`
	Status      int       `json:"status" bson:"status"`
	CreatedDate int64     `json:"created_date" bson:"created_date"`
}
