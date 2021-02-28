package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	domain "github.com/red-gold/telar-web/micros/notice/dto"
	models "github.com/red-gold/telar-web/micros/notice/models"
	service "github.com/red-gold/telar-web/micros/notice/services"
)

// UpdateNotificationHandle handle create a new notification
func UpdateNotificationHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create the model object
		var model models.NotificationModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		// Create service
		notificationService, serviceErr := service.NewNotificationService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		updatedNotification := &domain.Notification{
			ObjectId:             model.ObjectId,
			OwnerUserId:          req.UserID,
			OwnerDisplayName:     req.DisplayName,
			OwnerAvatar:          req.Avatar,
			Description:          model.Description,
			URL:                  model.URL,
			NotifyRecieverUserId: model.NotifyRecieverUserId,
			TargetId:             model.TargetId,
			IsSeen:               model.IsSeen,
			Type:                 model.Type,
			EmailNotification:    model.EmailNotification,
		}

		if err := notificationService.UpdateNotificationById(updatedNotification); err != nil {
			errorMessage := fmt.Sprintf("Update Notification Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("updateNotificationError", errorMessage)}, nil
		}
		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// SeenNotificationHandle handle create a new notification
func SeenNotificationHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// params from /notifications/seen/:notificationId
		notificationId := req.GetParamByName("notificationId")
		if notificationId == "" {
			errorMessage := fmt.Sprintf("Notification Id is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("notificationIdRequired", errorMessage)}, nil
		}
		notificationUUID, uuidErr := uuid.FromString(notificationId)
		if uuidErr != nil {
			errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("uuidError", errorMessage)}, nil
		}
		// Create service
		notificationService, serviceErr := service.NewNotificationService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		if err := notificationService.SeenNotification(notificationUUID, req.UserID); err != nil {
			errorMessage := fmt.Sprintf("Update Notification Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("updateNotificationError", errorMessage)}, nil
		}
		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}
