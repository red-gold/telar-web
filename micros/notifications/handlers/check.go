// Copyright (c) 2021 Amirhossein Movahedi (@qolzam)
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT
package handlers

import (
	"net/http"

	"github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	"github.com/red-gold/telar-core/pkg/log"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-web/micros/notifications/dto"
	service "github.com/red-gold/telar-web/micros/notifications/services"
)

// CheckNotifyEmailHandle handle query on notification
func CheckNotifyEmailHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		notificationService, serviceErr := service.NewNotificationService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		notificationList, err := notificationService.GetLastNotifications()
		if err != nil {
			log.Error("Get last notifications  - %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		if !(len(notificationList) > 0) {
			return handler.Response{
				StatusCode: http.StatusOK,
			}, nil
		}

		var recIds []uuid.UUID
		for _, notification := range notificationList {
			notification.IsEmailSent = true
			recIds = append(recIds, notification.NotifyRecieverUserId)
		}

		userInfoInReq := &UserInfoInReq{
			UserId:      req.UserID,
			Username:    req.Username,
			Avatar:      req.Avatar,
			DisplayName: req.DisplayName,
			SystemRole:  req.SystemRole,
		}
		mappedSettings, getSettingsErr := getUsersNotificationSettings(recIds, userInfoInReq)
		if err != nil {
			log.Error("Get users notification settings  - %s", getSettingsErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		var updateNotifyIds []uuid.UUID
		for _, notification := range notificationList {
			key := getSettingPath(notification.NotifyRecieverUserId, notificationSettingType, settingMappedFromNotify[notification.Type])
			if mappedSettings[key] == "true" {
				log.Info("Sending notify email to %s", notification.NotifyRecieverEmail)

				go func(notify dto.Notification) {
					err := sendEmailNotification(notify)
					if err != nil {
						log.Error("Send email notification - %s", err.Error())
					}
					log.Info("Notify email sent to %s", notify.NotifyRecieverEmail)
				}(notification)
			}

			updateNotifyIds = append(updateNotifyIds, notification.ObjectId)
		}

		if len(updateNotifyIds) > 0 {
			err = notificationService.UpdateEmailSent(updateNotifyIds)
			if err != nil {
				log.Error("Update last notifications  - %s", err.Error())
				return handler.Response{StatusCode: http.StatusInternalServerError}, err
			}
		}

		return handler.Response{
			StatusCode: http.StatusOK,
		}, nil
	}
}
