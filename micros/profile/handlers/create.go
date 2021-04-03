package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/profile/dto"
	service "github.com/red-gold/telar-web/micros/profile/services"
)

// InitProfileIndexHandle handle create a new index
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

// CreateProfileHandle handle create a new profile
func CreateDtoProfileHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create service
		profileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("Profile service Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("profileServiceError", errorMessage)}, nil
		}

		var model dto.UserProfile
		err := json.Unmarshal(req.Body, &model)
		if err != nil {
			errorMessage := fmt.Sprintf("Marshal user profile model %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("marshalUserProfileModel", errorMessage)}, nil
		}
		if err = profileService.SaveUserProfile(&model); err != nil {
			errorMessage := fmt.Sprintf("Create profile error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("createProfileError", errorMessage)}, nil
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true}`)),
			StatusCode: http.StatusOK,
		}, nil
	}
}
