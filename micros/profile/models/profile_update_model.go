package models

import (
	"github.com/red-gold/telar-web/constants"
)

type ProfileUpdateModel struct {
	FullName       string                        `json:"fullName" bson:"fullName"`
	Avatar         string                        `json:"avatar" bson:"avatar"`
	Banner         string                        `json:"banner" bson:"banner"`
	TagLine        string                        `json:"tagLine" bson:"tagLine"`
	Birthday       int64                         `json:"birthday" bson:"birthday"`
	WebUrl         string                        `json:"webUrl" bson:"webUrl"`
	CompanyName    string                        `json:"companyName" bson:"companyName"`
	FacebookId     string                        `json:"facebookId" bson:"facebookId"`
	InstagramId    string                        `json:"instagramId" bson:"instagramId"`
	TwitterId      string                        `json:"twitterId" bson:"twitterId"`
	LinkedInId     string                        `json:"linkedInId"`
	AccessUserList []string                      `json:"accessUserList" bson:"accessUserList"`
	Permission     constants.UserPermissionConst `json:"permission" bson:"permission"`
	LastUpdated    int64                         `json:"last_updated" bson:"last_updated"`
}

type ProfileGeneralUpdateModel struct {
	Address     string                        `json:"address" bson:"address"`
	Avatar      string                        `json:"avatar" bson:"avatar"`
	Banner      string                        `json:"banner" bson:"banner"`
	Country     string                        `json:"country" bson:"country"`
	FullName    string                        `json:"fullName" bson:"fullName"`
	SocialName  string                        `json:"socialName" bson:"socialName"`
	Permission  constants.UserPermissionConst `json:"permission" bson:"permission"`
	Phone       string                        `json:"phone" bson:"phone"`
	TagLine     string                        `json:"tagLine" bson:"tagLine"`
	LastUpdated int64                         `json:"last_updated" bson:"last_updated"`
}

type SocialInfoUpdateModel struct {
	FacebookId  string `json:"facebookId" bson:"facebookId"`
	InstagramId string `json:"instagramId" bson:"instagramId"`
	TwitterId   string `json:"twitterId" bson:"twitterId"`
	LinkedInId  string `json:"linkedInId" bson:"linkedInId"`
	LastUpdated int64  `json:"last_updated" bson:"last_updated"`
}
