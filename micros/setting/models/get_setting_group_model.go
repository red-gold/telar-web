package models

import uuid "github.com/gofrs/uuid"

type GetSettingGroupModel struct {
	Type        string                     `json:"type"`
	CreatedDate int64                      `json:"created_date"`
	OwnerUserId uuid.UUID                  `json:"ownerUserId"`
	List        []GetSettingGroupItemModel `json:"list"`
}
