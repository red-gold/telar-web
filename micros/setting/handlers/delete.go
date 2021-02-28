package handlers

import (
	"fmt"
	"net/http"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	service "github.com/red-gold/telar-web/micros/setting/services"
)

// DeleteUserSettingHandle handle delete a userSetting
func DeleteUserSettingHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// params from /userSettings/:userSettingId
		userSettingId := req.GetParamByName("userSettingId")
		if userSettingId == "" {
			errorMessage := fmt.Sprintf("UserSetting Id is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userSettingIdRequired", errorMessage)}, nil
		}
		fmt.Printf("\n UserSetting ID: %s", userSettingId)
		userSettingUUID, uuidErr := uuid.FromString(userSettingId)
		if uuidErr != nil {
			errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("uuidError", errorMessage)}, nil
		}
		fmt.Printf("\n UserSetting UUID: %s", userSettingUUID)
		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("UserSetting Service Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("userSettingServiceError", errorMessage)}, nil

		}

		if err := userSettingService.DeleteUserSettingByOwner(req.UserID, userSettingUUID); err != nil {
			errorMessage := fmt.Sprintf("Delete UserSetting Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("deleteUserSettingError", errorMessage)}, nil

		}
		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// DeleteUserAllSettingHandle handle delete all userSetting
func DeleteUserAllSettingHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("UserSetting Service Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("userSettingServiceError", errorMessage)}, nil

		}

		if err := userSettingService.DeleteUserSettingByOwnerUserId(req.UserID); err != nil {
			errorMessage := fmt.Sprintf("Delete UserSetting Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("deleteUserSettingError", errorMessage)}, nil

		}
		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}
