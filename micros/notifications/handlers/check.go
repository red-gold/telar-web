// Copyright (c) 2021 Amirhossein Movahedi (@qolzam)
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT
package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/notifications/database"
	"github.com/red-gold/telar-web/micros/notifications/dto"
	service "github.com/red-gold/telar-web/micros/notifications/services"
)

// CheckNotifyEmailHandle handle query on notification
func CheckNotifyEmailHandle(c *fiber.Ctx) error {

	// Create service
	notificationService, serviceErr := service.NewNotificationService(database.Db)
	if serviceErr != nil {
		log.Error("[CheckNotifyEmailHandle.NewNotificationService] %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/notificationService", "Error happened while creating notificationService!"))
	}

	notificationList, err := notificationService.GetLastNotifications()
	if err != nil {
		log.Error("[CheckNotifyEmailHandle.GetLastNotifications] %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/notificationList", "Error happened while getting notification list!"))
	}

	if !(len(notificationList) > 0) {
		return c.SendStatus(http.StatusOK)
	}

	var recIds []uuid.UUID
	for _, notification := range notificationList {
		notification.IsEmailSent = true
		recIds = append(recIds, notification.NotifyRecieverUserId)
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		currentUser = types.UserContext{}
	}

	userInfoInReq := &UserInfoInReq{
		UserId:      currentUser.UserID,
		Username:    currentUser.Username,
		Avatar:      currentUser.Avatar,
		DisplayName: currentUser.DisplayName,
		SystemRole:  currentUser.SystemRole,
	}
	mappedSettings, getSettingsErr := getUsersNotificationSettings(recIds, userInfoInReq)
	if err != nil {
		log.Error("[CheckNotifyEmailHandle.getUsersNotificationSettings] %s", getSettingsErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/notificationSettings", "Error happened while getting user notification setting!"))

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
			log.Error("[CheckNotifyEmailHandle.UpdateEmailSent] %s", getSettingsErr.Error())
			return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/updateEmailSent", "Error happened while updating notification!"))
		}
	}

	return c.SendStatus(http.StatusOK)

}
