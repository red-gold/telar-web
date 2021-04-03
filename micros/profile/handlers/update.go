package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	models "github.com/red-gold/telar-web/micros/profile/models"
	service "github.com/red-gold/telar-web/micros/profile/services"
)

// UpdateProfileHandle a function invocation
func UpdateProfileHandle(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {

		// Create service
		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		model := models.ProfileUpdateModel{}
		unmarshalErr := json.Unmarshal(req.Body, &model)
		if unmarshalErr != nil {
			errorMessage := fmt.Sprintf("Error while un-marshaling ProfileUpdateModel: %s",
				unmarshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("marshalProfileUpdateModelError", errorMessage)},
				unmarshalErr

		}
		fmt.Printf("Update profile userId: %s username: %s", req.UserID, req.Username)
		foundUserProfile, profileErr := userProfileService.FindByUserId(req.UserID)
		if foundUserProfile == nil {
			return handler.Response{
				StatusCode: http.StatusNotFound,
				Body:       utils.MarshalError("notFoundUser", "Could not find user "+req.UserID.String()),
			}, nil
		}
		if profileErr != nil {
			errorMessage := fmt.Sprintf("Find profile Error %s",
				profileErr.Error())

			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("findProfileError", errorMessage)},
				profileErr
		}
		foundUserProfile.FullName = model.FullName
		foundUserProfile.Avatar = model.Avatar
		foundUserProfile.Banner = model.Banner
		foundUserProfile.TagLine = model.TagLine
		foundUserProfile.Birthday = model.Birthday
		foundUserProfile.WebUrl = model.WebUrl
		foundUserProfile.CompanyName = model.CompanyName
		foundUserProfile.FacebookId = model.FacebookId
		foundUserProfile.InstagramId = model.InstagramId
		foundUserProfile.TwitterId = model.TwitterId
		foundUserProfile.AccessUserList = model.AccessUserList
		foundUserProfile.Permission = model.Permission

		err := userProfileService.UpdateUserProfileById(req.UserID, foundUserProfile)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true}`)),
			StatusCode: http.StatusOK,
		}, nil
	}

}

// UpdateLastSeen a function invocation
func UpdateLastSeen(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		var model models.UpdateLastSeenModel

		unmarshalErr := json.Unmarshal(req.Body, &model)
		if unmarshalErr != nil {
			errorMessage := fmt.Sprintf("Unmarshal models.UpdateLastSeenModel %s",
				unmarshalErr.Error())
			println(errorMessage)
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("unmarshalUpdateLastSeenError", errorMessage)},
				unmarshalErr
		}

		// Create service
		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		err := userProfileService.UpdateLastSeenNow(model.UserId)
		if err != nil {
			errorMessage := fmt.Sprintf("Update last seen %s",
				err.Error())
			println(errorMessage)
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("updateLastSeenError", errorMessage)},
				err
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true}`)),
			StatusCode: http.StatusOK,
		}, nil
	}

}
