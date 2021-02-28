package models

import (
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-web/constants"
)

type MyProfileModel struct {
	ObjectId       uuid.UUID                     `json:"objectId"`
	FullName       string                        `json:"fullName"`
	Avatar         string                        `json:"avatar"`
	Banner         string                        `json:"banner"`
	TagLine        string                        `json:"tagLine"`
	Birthday       int64                         `json:"birthday"`
	WebUrl         string                        `json:"webUrl"`
	CompanyName    string                        `json:"companyName"`
	FacebookId     string                        `json:"facebookId"`
	InstagramId    string                        `json:"instagramId"`
	TwitterId      string                        `json:"twitterId"`
	AccessUserList []string                      `json:"accessUserList"`
	Permission     constants.UserPermissionConst `json:"permission"`
}
