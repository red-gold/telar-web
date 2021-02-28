package handlers

import (
	"fmt"
	"net/http"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	service "github.com/red-gold/telar-web/micros/notifications/services"
)

// DeleteNotificationHandle handle delete a Notification
func DeleteNotificationHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// params from /notifications/id/:notificationId
		notificationId := req.GetParamByName("notificationId")
		if notificationId == "" {
			errorMessage := fmt.Sprintf("Notification Id is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("notificationIdRequired", errorMessage)}, nil
		}
		fmt.Printf("\n Notification ID: %s", notificationId)
		notificationUUID, uuidErr := uuid.FromString(notificationId)
		if uuidErr != nil {
			errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("uuidError", errorMessage)}, nil
		}
		fmt.Printf("\n Notification UUID: %s", notificationUUID)
		// Create service
		notificationService, serviceErr := service.NewNotificationService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("Notification Service Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("notificationServiceError", errorMessage)}, nil

		}

		if err := notificationService.DeleteNotificationByOwner(req.UserID, notificationUUID); err != nil {
			errorMessage := fmt.Sprintf("Delete Notification Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("deleteNotificationError", errorMessage)}, nil

		}
		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// DeleteNotificationByUserIdHandle handle delete a Notification but userId
func DeleteNotificationByUserIdHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create service
		notificationService, serviceErr := service.NewNotificationService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("Notification Service Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("NotificationServiceError", errorMessage)}, nil

		}

		if err := notificationService.DeleteNotificationsByUserId(req.UserID); err != nil {
			errorMessage := fmt.Sprintf("Delete Notification Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("deleteNotificationError", errorMessage)}, nil
		}

		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}
