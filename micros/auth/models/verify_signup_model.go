package models

type VerifySignupModel struct {
	Code         string `json:"code"`
	Token        string `json:"verificaitonSecret"`
	ResponseType string `json:"responseType"`
}
