package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/notifications/database"
	models "github.com/red-gold/telar-web/micros/notifications/models"
	service "github.com/red-gold/telar-web/micros/notifications/services"
)

// GetNotificationsByUserIdHandle handle query on notification
func GetNotificationsByUserIdHandle(c *fiber.Ctx) error {

	// Create service
	notificationService, serviceErr := service.NewNotificationService(database.Db)
	if serviceErr != nil {
		log.Error("[GetNotificationsByUserIdHandle.NewNotificationService] %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/notificationService", "Error happened while creating notificationService!"))
	}

	pageParam := c.Query("page")

	page := 0
	if pageParam != "" {
		var strErr error
		page, strErr = strconv.Atoi(pageParam)
		if strErr != nil {
			log.Error("[GetNotificationsByUserIdHandle.strconv.Atoi] %s", strErr.Error())
			return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal", "Error happened while parsing page!"))
		}
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[GetNotificationsByUserIdHandle] Can not get current user")
		return c.Status(http.StatusUnauthorized).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	notificationList, err := notificationService.GetNotificationByUserId(&currentUser.UserID, "created_date", int64(page))
	if err != nil {
		log.Error("[GetNotificationsByUserIdHandle.GetNotificationByUserId] %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/getNotificationByUserId", "Error happened while reading notification!"))
	}

	return c.JSON(notificationList)

}

// GetNotificationHandle handle get a notification
func GetNotificationHandle(c *fiber.Ctx) error {

	// Create service
	notificationService, serviceErr := service.NewNotificationService(database.Db)
	if serviceErr != nil {
		log.Error("[GetNotificationHandle.NewNotificationService] %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/notificationService", "Error happened while creating notificationService!"))
	}
	notificationId := c.Params("notificationId")
	notificationUUID, uuidErr := uuid.FromString(notificationId)
	if uuidErr != nil {
		errorMessage := fmt.Sprintf("Notification Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("notificationIdRequired", "Notification id is required!"))

	}

	foundNotification, err := notificationService.FindById(notificationUUID)
	if err != nil {
		log.Error("[GetNotificationHandle.notificationService.FindById] %s - %s", notificationUUID.String(), serviceErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findNotification", "Error happened while finding notification!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[GetNotificationHandle] Can not get current user")
		return c.Status(http.StatusUnauthorized).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	notificationModel := models.NotificationModel{
		ObjectId:             foundNotification.ObjectId,
		OwnerUserId:          currentUser.UserID,
		OwnerDisplayName:     currentUser.DisplayName,
		OwnerAvatar:          currentUser.Avatar,
		Description:          foundNotification.Description,
		URL:                  foundNotification.URL,
		NotifyRecieverUserId: foundNotification.NotifyRecieverUserId,
		TargetId:             foundNotification.TargetId,
		IsSeen:               foundNotification.IsSeen,
		Type:                 foundNotification.Type,
		EmailNotification:    foundNotification.EmailNotification,
	}

	return c.JSON(notificationModel)

}
