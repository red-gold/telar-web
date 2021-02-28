package models

type LoginModel struct {
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password" `
}
