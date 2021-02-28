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

// UpdateUserSettingHandle handle create a new userSetting
func UpdateUserSettingHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create the model object
		var model models.UpdateSettingGroupModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		if model.Type == "" {
			errorMessage := fmt.Sprintf("Setting type can not be empty Error")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("settingTypeEmptyError", errorMessage)}, nil
		}

		var userSettings []domain.UserSetting
		for _, setting := range model.List {

			updatedUserSetting := domain.UserSetting{
				ObjectId:    setting.ObjectId,
				OwnerUserId: req.UserID,
				CreatedDate: utils.UTCNowUnix(),
				Name:        setting.Name,
				Value:       setting.Value,
				Type:        model.Type,
				IsSystem:    false,
			}
			userSettings = append(userSettings, updatedUserSetting)
		}

		if !(len(userSettings) > 0) {
			errorMessage := fmt.Sprintf("No setting added for update Error")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("noSettingForUpdateError", errorMessage)}, nil
		}
		if err := userSettingService.UpdateUserSettingsById(req.UserID, userSettings); err != nil {
			errorMessage := fmt.Sprintf("Update UserSetting Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("updateUserSettingError", errorMessage)}, nil
		}
		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}
