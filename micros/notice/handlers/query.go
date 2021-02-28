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
	models "github.com/red-gold/telar-web/micros/notice/models"
	service "github.com/red-gold/telar-web/micros/notice/services"
)

// GetNotificationsByUserIdHandle handle query on notification
func GetNotificationsByUserIdHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		notificationService, serviceErr := service.NewNotificationService(db)
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
		pageParam := query.Get("page")

		page := 0
		if pageParam != "" {
			var strErr error
			page, strErr = strconv.Atoi(pageParam)
			if strErr != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, strErr
			}
		}

		notificationList, err := notificationService.GetNotificationByUserId(&req.UserID, "created_date", int64(page))
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		body, marshalErr := json.Marshal(notificationList)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("Error while marshaling notificationList: %s",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("notificationListMarshalError", errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// GetNotificationHandle handle get a notification
func GetNotificationHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		notificationService, serviceErr := service.NewNotificationService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}
		notificationId := req.GetParamByName("notificationId")
		notificationUUID, uuidErr := uuid.FromString(notificationId)
		if uuidErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("parseUUIDError", "Can not parse notification id!")},
				nil
		}

		foundNotification, err := notificationService.FindById(notificationUUID)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		notificationModel := models.NotificationModel{
			ObjectId:             foundNotification.ObjectId,
			OwnerUserId:          req.UserID,
			OwnerDisplayName:     req.DisplayName,
			OwnerAvatar:          req.Avatar,
			Description:          foundNotification.Description,
			URL:                  foundNotification.URL,
			NotifyRecieverUserId: foundNotification.NotifyRecieverUserId,
			TargetId:             foundNotification.TargetId,
			IsSeen:               foundNotification.IsSeen,
			Type:                 foundNotification.Type,
			EmailNotification:    foundNotification.EmailNotification,
		}

		body, marshalErr := json.Marshal(notificationModel)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("{error: 'Error while marshaling notificationModel: %s'}",
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
