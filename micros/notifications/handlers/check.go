// Copyright (c) 2021 Amirhossein Movahedi (@qolzam)
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT
package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	coreConfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	notifyConfig "github.com/red-gold/telar-web/micros/notifications/config"
	"github.com/red-gold/telar-web/micros/notifications/database"
	"github.com/red-gold/telar-web/micros/notifications/dto"
	service "github.com/red-gold/telar-web/micros/notifications/services"
	"github.com/valyala/bytebufferpool"
)

// CheckNotifyEmailHandle godoc
// @Summary Check and send notification emails
// @Description Checks the latest notifications and sends emails to the users if necessary
// @Tags Notification
// @Accept json
// @Produce json
// @Security JWT
// @Param Authorization header string true "Authentication" default(Bearer <Add_token_here>)
// @Success 200 {string} string "OK"
// @Failure 500 {object} utils.TelarError
// @Router /check [post]
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

			go func(notify dto.Notification, c *fiber.Ctx) {
				buf := bytebufferpool.Get()
				defer bytebufferpool.Put(buf)
				notify.Title = getNotificationTitleByType(notify.Type, notify.OwnerDisplayName)
				emailData := fiber.Map{

					"AppName":         *coreConfig.AppConfig.AppName,
					"AppURL":          notifyConfig.NotificationConfig.WebURL,
					"Title":           notify.Title,
					"Avatar":          notify.OwnerAvatar,
					"FullName":        notify.OwnerDisplayName,
					"ViewLink":        combineURL(notifyConfig.NotificationConfig.WebURL, notify.URL),
					"UnsubscribeLink": combineURL(notifyConfig.NotificationConfig.WebURL, "settings/notify"),
				}
				c.App().Config().Views.Render(buf, "notify_email", emailData, c.App().Config().ViewsLayout)
				err := sendEmailNotification(notify, buf.String())
				if err != nil {
					log.Error("Send email notification - %s", err.Error())
				}
				log.Info("Notify email sent to %s", notify.NotifyRecieverEmail)
			}(notification, c)
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
