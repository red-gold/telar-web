package models

type CreateVerifyTokenTokenModel struct {
	Code      string `json:"code"`
	Token     string `json:"token"`
	Recaptcha string `json:"g-recaptcha-response"`
}
