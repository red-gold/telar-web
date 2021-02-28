package handlers

import (
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

// InitProfileIndexHandle handle create a new post
func InitProfileIndexHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create service
		profileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("Profile service Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("profileServiceError", errorMessage)}, nil
		}

		postIndexMap := make(map[string]interface{})
		postIndexMap["fullName"] = "text"
		postIndexMap["objectId"] = 1
		if err := profileService.CreateUserProfileIndex(postIndexMap); err != nil {
			errorMessage := fmt.Sprintf("Create post index Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("createPostIndexError", errorMessage)}, nil
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true}`)),
			StatusCode: http.StatusOK,
		}, nil
	}
}
