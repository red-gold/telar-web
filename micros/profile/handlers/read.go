package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	"github.com/red-gold/telar-core/pkg/log"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	models "github.com/red-gold/telar-web/micros/profile/models"
	service "github.com/red-gold/telar-web/micros/profile/services"
)

type MembersPayload struct {
	Users map[string]interface{} `json:"users"`
}

// ReadDtoProfileHandle a function invocation
func ReadDtoProfileHandle(db interface{}) func(server.Request) (handler.Response, error) {

	log.Info("ReadDtoProfileHandle")
	return func(req server.Request) (handler.Response, error) {
		userId := req.GetParamByName("userId")
		userUUID, uuidErr := uuid.FromString(userId)
		if uuidErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("parseUUIDError", "Can not parse user id!")},
				nil
		}
		// Create service
		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		foundUser, err := userProfileService.FindByUserId(userUUID)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		if foundUser == nil {
			return handler.Response{
				StatusCode: http.StatusNotFound,
				Body:       utils.MarshalError("notFoundUser", "Could not find user "+req.UserID.String()),
			}, nil
		}

		body, marshalErr := json.Marshal(foundUser)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("{error: 'Error while marshaling userProfile: %s'}",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: []byte(errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// ReadProfileHandle a function invocation
func ReadProfileHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		userId := req.GetParamByName("userId")
		userUUID, uuidErr := uuid.FromString(userId)
		if uuidErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("parseUUIDError", "Can not parse user id!")},
				nil
		}
		// Create service
		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		foundUser, err := userProfileService.FindByUserId(userUUID)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		if foundUser == nil {
			return handler.Response{
				StatusCode: http.StatusNotFound,
				Body:       utils.MarshalError("notFoundUser", "Could not find user "+req.UserID.String()),
			}, nil
		}

		profileModel := models.MyProfileModel{
			ObjectId:       foundUser.ObjectId,
			FullName:       foundUser.FullName,
			Avatar:         foundUser.Avatar,
			Banner:         foundUser.Banner,
			TagLine:        foundUser.TagLine,
			Birthday:       foundUser.Birthday,
			WebUrl:         foundUser.WebUrl,
			CompanyName:    foundUser.CompanyName,
			FacebookId:     foundUser.FacebookId,
			InstagramId:    foundUser.InstagramId,
			TwitterId:      foundUser.TwitterId,
			AccessUserList: foundUser.AccessUserList,
			Permission:     foundUser.Permission,
		}

		body, marshalErr := json.Marshal(profileModel)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("{error: 'Error while marshaling userProfile: %s'}",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: []byte(errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// ReadMyProfileHandle a function invocation to read authed user profile
func ReadMyProfileHandle(db interface{}) func(server.Request) (handler.Response, error) {
	fmt.Printf("\n[INFO] FOUND USER main ReadMyProfileHandle")

	return func(req server.Request) (handler.Response, error) {
		// Create service
		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		foundUser, err := userProfileService.FindByUserId(req.UserID)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}
		if foundUser == nil {
			return handler.Response{
				StatusCode: http.StatusNotFound,
				Body:       utils.MarshalError("notFoundUser", "Could not find user "+req.UserID.String()),
			}, nil
		}

		profileModel := models.MyProfileModel{
			ObjectId:       foundUser.ObjectId,
			FullName:       foundUser.FullName,
			Avatar:         foundUser.Avatar,
			Banner:         foundUser.Banner,
			TagLine:        foundUser.TagLine,
			Birthday:       foundUser.Birthday,
			WebUrl:         foundUser.WebUrl,
			CompanyName:    foundUser.CompanyName,
			FacebookId:     foundUser.FacebookId,
			InstagramId:    foundUser.InstagramId,
			TwitterId:      foundUser.TwitterId,
			AccessUserList: foundUser.AccessUserList,
			Permission:     foundUser.Permission,
		}
		body, marshalErr := json.Marshal(profileModel)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("Error while marshaling userProfile: %s",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("marshalUserProfileError", errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// DispatchProfilesHandle a function invocation to read authed user profile
func DispatchProfilesHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Parse model object
		var model models.DispatchProfilesModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			errorMessage := fmt.Sprintf("Unmarshal  models.DispatchProfilesModel array %s", err.Error())
			println(errorMessage)
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("dispatchProfilesModelMarshalError", errorMessage)}, nil
		}

		// Create service
		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		foundUsers, err := userProfileService.FindProfileByUserIds(model.UserIds)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		mappedUsers := make(map[string]interface{})
		for _, v := range foundUsers {
			mappedUser := make(map[string]interface{})
			mappedUser["userId"] = v.ObjectId
			mappedUser["fullName"] = v.FullName
			mappedUser["avatar"] = v.Avatar
			mappedUser["banner"] = v.Banner
			mappedUser["tagLine"] = v.TagLine
			mappedUser["lastSeen"] = v.LastSeen
			mappedUser["createdDate"] = v.CreatedDate
			mappedUsers[v.ObjectId.String()] = mappedUser
		}

		actionRoomPayload := &MembersPayload{
			Users: mappedUsers,
		}

		activeRoomAction := Action{
			Type:    "SET_USER_ENTITIES",
			Payload: actionRoomPayload,
		}

		userInfoReq := &UserInfoInReq{
			UserId:      req.UserID,
			Username:    req.Username,
			Avatar:      req.Avatar,
			DisplayName: req.DisplayName,
			SystemRole:  req.SystemRole,
		}

		go dispatchAction(activeRoomAction, userInfoReq)

		return handler.Response{
			StatusCode: http.StatusOK,
		}, nil
	}
}

// GetProfileByIds a function invocation to profiles by ids
func GetProfileByIds(db interface{}) func(server.Request) (handler.Response, error) {
	return func(req server.Request) (handler.Response, error) {

		// Parse model object
		var model models.GetProfilesModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			errorMessage := fmt.Sprintf("Unmarshal  models.GetProfilesModel array %s", err.Error())
			println(errorMessage)
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("getProfilesModelMarshalError", errorMessage)}, nil
		}

		// Create service
		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		foundUsers, err := userProfileService.FindProfileByUserIds(model.UserIds)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		body, marshalErr := json.Marshal(foundUsers)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("Error while marshaling userProfiles: %s",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("marshalUserProfilesError", errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       body,
			StatusCode: http.StatusOK,
		}, nil
	}
}
