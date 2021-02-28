package models

import (
	uuid "github.com/gofrs/uuid"
)

type ActionRoomModel struct {
	ObjectId    uuid.UUID `json:"objectId"`
	OwnerUserId uuid.UUID `json:"ownerUserId"`
	PrivateKey  string    `json:"privateKey"`
	AccessKey   string    `json:"accessKey"`
	Status      int       `json:"status"`
	CreatedDate int64     `json:"created_date"`
}
