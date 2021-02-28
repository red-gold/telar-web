package dto

import (
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-web/constants"
)

type UserVerification struct {
	ObjectId        uuid.UUID             `json:"objectId" bson:"objectId"`
	Code            string                `json:"code" bson:"code"`
	Target          string                `json:"target" bson:"target"`
	TargetType      constants.VerifyConst `json:"targetType" bson:"targetType"`
	Counter         int64                 `json:"counter" bson:"counter"`
	CreatedDate     int64                 `json:"created_date" bson:"created_date"`
	RemoteIpAddress string                `json:"remoteIpAddress" bson:"remoteIpAddress"`
	UserId          uuid.UUID             `json:"userId" bson:"userId"`
	IsVerified      bool                  `json:"isVerified" bson:"isVerified"`
	LastUpdated     int64                 `json:"last_updated" bson:"last_updated"`
}
