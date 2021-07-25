package models

type ProfileUpdateModel struct {
	FullName   string `json:"fullName" bson:"fullName"`
	Avatar     string `json:"avatar" bson:"avatar"`
	Banner     string `json:"banner" bson:"banner"`
	TagLine    string `json:"tagLine" bson:"tagLine"`
	SocialName string `json:"socialName" bson:"socialName"`
}
