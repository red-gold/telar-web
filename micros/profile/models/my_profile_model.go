package models

import (
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-web/constants"
)

type Location struct {
	GeoJSONType string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type MyProfileModel struct {
	ObjectId       uuid.UUID                     `json:"objectId"`
	FullName       string                        `json:"fullName"`
	SocialName     string                        `json:"socialName"`
	Avatar         string                        `json:"avatar"`
	Banner         string                        `json:"banner"`
	TagLine        string                        `json:"tagLine"`
	CreatedDate    int64                         `json:"created_date"`
	LastUpdated    int64                         `json:"last_updated"`
	LastSeen       int64                         `json:"lastSeen"`
	Email          string                        `json:"email"`
	Birthday       int64                         `json:"birthday"`
	WebUrl         string                        `json:"webUrl"`
	CompanyName    string                        `json:"companyName"`
	Country        string                        `json:"country"`
	Address        string                        `json:"address"`
	Phone          string                        `json:"phone"`
	VoteCount      int64                         `json:"voteCount"`
	ShareCount     int64                         `json:"shareCount"`
	FollowCount    int64                         `json:"followCount"`
	FollowerCount  int64                         `json:"followerCount"`
	PostCount      int64                         `json:"postCount"`
	FacebookId     string                        `json:"facebookId"`
	InstagramId    string                        `json:"instagramId"`
	TwitterId      string                        `json:"twitterId"`
	LinkedInId     string                        `json:"linkedInId"`
	AccessUserList []string                      `json:"accessUserList"`
	Permission     constants.UserPermissionConst `json:"permission"`
}
