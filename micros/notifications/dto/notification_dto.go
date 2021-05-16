package dto

import (
	uuid "github.com/gofrs/uuid"
)

type Notification struct {
	ObjectId             uuid.UUID `json:"objectId" bson:"objectId"`
	OwnerUserId          uuid.UUID `json:"ownerUserId" bson:"ownerUserId"`
	OwnerDisplayName     string    `json:"ownerDisplayName" bson:"ownerDisplayName"`
	OwnerAvatar          string    `json:"ownerAvatar" bson:"ownerAvatar"`
	CreatedDate          int64     `json:"created_date" bson:"created_date"`
	Description          string    `json:"description" bson:"description"`
	URL                  string    `json:"url" bson:"url"`
	NotifyRecieverUserId uuid.UUID `json:"notifyRecieverUserId" bson:"notifyRecieverUserId"`
	NotifyRecieverEmail  string    `json:"notifyRecieverEmail" bson:"notifyRecieverEmail"`
	TargetId             uuid.UUID `json:"targetId" bson:"targetId"`
	IsSeen               bool      `json:"isSeen" bson:"isSeen"`
	Type                 string    `json:"type" bson:"type"`
	EmailNotification    int16     `json:"emailNotification" bson:"emailNotification"`
	IsEmailSent          bool      `json:"isEmailSent" bson:"isEmailSent"`
}
