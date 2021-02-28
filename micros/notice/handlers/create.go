package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	domain "github.com/red-gold/telar-web/micros/notice/dto"
	models "github.com/red-gold/telar-web/micros/notice/models"
	service "github.com/red-gold/telar-web/micros/notice/services"
)

type NotificationAction struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// CreateNotificationHandle handle create a new notification
func CreateNotificationHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create the model object
		var model models.CreateNotificationModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			errorMessage := fmt.Sprintf("Unmarshal CreateNotificationModel Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("modelMarshalError", errorMessage)}, nil
		}

		// Create service
		notificationService, serviceErr := service.NewNotificationService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("notification Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("notificationServiceError", errorMessage)}, nil
		}

		newNotification := &domain.Notification{
			ObjectId:             model.ObjectId,
			OwnerUserId:          req.UserID,
			OwnerDisplayName:     req.DisplayName,
			OwnerAvatar:          req.Avatar,
			Description:          model.Description,
			URL:                  model.URL,
			NotifyRecieverUserId: model.NotifyRecieverUserId,
			TargetId:             model.TargetId,
			IsSeen:               false,
			Type:                 model.Type,
			EmailNotification:    model.EmailNotification,
		}
		if err := notificationService.SaveNotification(newNotification); err != nil {
			errorMessage := fmt.Sprintf("Save Notification Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("saveNotificationError", errorMessage)}, nil
		}

		// Send notification
		go func() {

			actionURL := fmt.Sprintf("/actions/dispatch/%s", model.NotifyRecieverUserId.String())

			notificationList := make(map[string]domain.Notification)
			notificationList[newNotification.ObjectId.String()] = *newNotification
			notificationAction := &NotificationAction{
				Type:    "ADD_PLAIN_NOTIFY_LIST",
				Payload: notificationList,
			}

			notificationActionBytes, marshalErr := json.Marshal(notificationAction)
			if marshalErr != nil {
				errorMessage := fmt.Sprintf("Marshal notification Error %s", marshalErr.Error())
				fmt.Println(errorMessage)
			}
			// Create user headers for http request
			userHeaders := make(map[string][]string)
			userHeaders["uid"] = []string{req.UserID.String()}
			userHeaders["email"] = []string{req.Username}
			userHeaders["avatar"] = []string{req.Avatar}
			userHeaders["displayName"] = []string{req.DisplayName}
			userHeaders["role"] = []string{req.SystemRole}

			_, actionErr := functionCall(http.MethodPost, notificationActionBytes, actionURL, userHeaders)

			if actionErr != nil {
				errorMessage := fmt.Sprintf("Cannot send action request! error: %s", actionErr.Error())
				fmt.Println(errorMessage)
			}
		}()

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true, "objectId": "%s"}`, newNotification.ObjectId.String())),
			StatusCode: http.StatusOK,
		}, nil
	}
}
