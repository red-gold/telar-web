package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	models "github.com/red-gold/telar-web/micros/setting/models"
	service "github.com/red-gold/telar-web/micros/setting/services"
)

// QueryUserSettingHandle handle quey on userSetting
func QueryUserSettingHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		var query *url.Values
		if len(req.QueryString) > 0 {
			q, err := url.ParseQuery(string(req.QueryString))
			if err != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, err
			}
			query = &q
		}
		searchParam := query.Get("search")
		pageParam := query.Get("page")
		ownerUserIdParam := query.Get("owner")
		userSettingTypeIdParam := query.Get("type")

		var ownerUserId *uuid.UUID = nil
		if ownerUserIdParam != "" {

			parsedUUID, uuidErr := uuid.FromString(ownerUserIdParam)

			if uuidErr != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, uuidErr
			}

			ownerUserId = &parsedUUID
		}

		var userSettingTypeId *int = nil
		if userSettingTypeIdParam != "" {

			parsedType, strErr := strconv.Atoi(userSettingTypeIdParam)
			if strErr != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, strErr
			}
			userSettingTypeId = &parsedType
		}
		page := 0
		if pageParam != "" {
			var strErr error
			page, strErr = strconv.Atoi(pageParam)
			if strErr != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, strErr
			}
		}
		userSettingList, err := userSettingService.QueryUserSetting(searchParam, ownerUserId, userSettingTypeId, "created_date", int64(page))
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		body, marshalErr := json.Marshal(userSettingList)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("Error while marshaling userSettingList: %s",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userSettingListMarshalError", errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// GetAllUserSetting handle get all userSetting
func GetAllUserSetting(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		userSettingList, err := userSettingService.GetAllUserSetting(req.UserID)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		groupSettingsMap := make(map[string][]models.GetSettingGroupItemModel)
		for _, setting := range userSettingList {

			settingModel := models.GetSettingGroupItemModel{
				ObjectId: setting.ObjectId,
				Name:     setting.Name,
				Value:    setting.Value,
				IsSystem: setting.IsSystem,
			}
			groupSettingsMap[setting.Type] = append(groupSettingsMap[setting.Type], settingModel)
		}

		body, marshalErr := json.Marshal(groupSettingsMap)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("Error while marshaling userSettingList: %s",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userSettingListMarshalError", errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// GetAllUserSettingByType handle get all userSetting
func GetAllUserSettingByType(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		settingType := req.GetParamByName("type")
		if settingType == "" {
			errorMessage := fmt.Sprintf("Error setting type can not be empty.")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("settingTypeEmptyError", errorMessage)}, nil
		}
		userSettingList, err := userSettingService.GetAllUserSettingByType(req.UserID, settingType)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		var groupSettingItems []models.GetSettingGroupItemModel
		for _, setting := range userSettingList {

			settingModel := models.GetSettingGroupItemModel{
				ObjectId: setting.ObjectId,
				Name:     setting.Name,
				Value:    setting.Value,
				IsSystem: setting.IsSystem,
			}
			groupSettingItems = append(groupSettingItems, settingModel)
		}
		groupSettingsModel := models.GetSettingGroupModel{
			Type:        settingType,
			CreatedDate: userSettingList[0].CreatedDate,
			OwnerUserId: req.UserID,
			List:        groupSettingItems,
		}

		body, marshalErr := json.Marshal(groupSettingsModel)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("Error while marshaling userSettingList: %s",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userSettingListMarshalError", errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// GetUserSettingHandle handle get userSetting
func GetUserSettingHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		userSettingService, serviceErr := service.NewUserSettingService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}
		userSettingId := req.GetParamByName("userSettingId")
		userSettingUUID, uuidErr := uuid.FromString(userSettingId)
		if uuidErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("parseUUIDError", "Can not parse userSetting id!")},
				nil
		}

		foundUserSetting, err := userSettingService.FindById(userSettingUUID)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		userSettingModel := models.UserSettingModel{
			ObjectId:    foundUserSetting.ObjectId,
			OwnerUserId: req.UserID,
			CreatedDate: foundUserSetting.CreatedDate,
			Name:        foundUserSetting.Name,
			Value:       foundUserSetting.Value,
			Type:        foundUserSetting.Type,
			IsSystem:    foundUserSetting.IsSystem,
		}

		body, marshalErr := json.Marshal(userSettingModel)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("{error: 'Error while marshaling userSettingModel: %s'}",
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
