package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	domain "github.com/red-gold/telar-web/micros/setting/dto"
	models "github.com/red-gold/telar-web/micros/setting/models"
	service "github.com/red-gold/telar-web/micros/setting/services"
)

// CreateUserSettingHandle handle create a new userSetting
func CreateUserSettingHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create the model object
		var model models.CreateUserSettingModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			errorMessage := fmt.Sprintf("Unmarshal CreateUserSettingModel Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("modelMarshalError", errorMessage)}, nil
		}

		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("userSetting Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("userSettingServiceError", errorMessage)}, nil
		}
		newUserSetting := &domain.UserSetting{
			OwnerUserId: req.UserID,
			Name:        model.Name,
			Value:       model.Value,
			Type:        model.Type,
			IsSystem:    false,
		}

		if err := userSettingService.SaveUserSetting(newUserSetting); err != nil {
			errorMessage := fmt.Sprintf("Save UserSetting Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("saveUserSettingError", errorMessage)}, nil
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true, "objectId": "%s"}`, newUserSetting.ObjectId.String())),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// CreateSettingGroupHandle handle create a new userSetting
func CreateSettingGroupHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create the model object
		var model models.CreateSettingGroupModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			errorMessage := fmt.Sprintf("Unmarshal CreateSettingGroupModel Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("modelMarshalError", errorMessage)}, nil
		}

		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("userSetting Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("userSettingServiceError", errorMessage)}, nil
		}

		var userSettingList []domain.UserSetting
		for _, setting := range model.List {

			newUserSetting := domain.UserSetting{
				OwnerUserId: req.UserID,
				Name:        setting.Name,
				Value:       setting.Value,
				Type:        model.Type,
				IsSystem:    false,
			}
			userSettingList = append(userSettingList, newUserSetting)
		}

		if err := userSettingService.SaveManyUserSetting(userSettingList); err != nil {
			errorMessage := fmt.Sprintf("Save UserSetting Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("saveUserSettingError", errorMessage)}, nil
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true}`)),
			StatusCode: http.StatusOK,
		}, nil
	}
}
