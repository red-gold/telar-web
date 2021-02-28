package models

import "fmt"

// UserAuthError is a custom error from Go built-in error
type UserAuthError struct {
	Code string
}

const (
	UserAuthErrorUserNotVerified = "Auth/usersNotVerifide"
)

// Error get message by error code
func (e UserAuthError) Error() string {
	switch e.Code {
	case UserAuthErrorUserNotVerified:
		return "User is not verified!"
	default:
		return "Unrecognized user error code"
	}
}

// ErrorResponseF get error in json format for api response
func (e UserAuthError) ErrorResponse() []byte {
	result := fmt.Sprintf(`{code: "%s", message: "%s"}`, e.Code, e.Error())
	return []byte(result)
}
