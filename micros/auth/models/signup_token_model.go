package models

type SignupTokenModel struct {
	User         UserSignupTokenModel `json:"user"`
	VerifyType   string               `json:"verifyType"`
	Recaptcha    string               `json:"g-recaptcha-response"`
	ResponseType string               `json:"responseType"`
}

type UserSignupTokenModel struct {
	Fullname string `json:"fullName"`
	Email    string `json:"email" `
	Password string `json:"password" `
}
