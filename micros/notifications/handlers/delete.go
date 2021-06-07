package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/notifications/database"
	service "github.com/red-gold/telar-web/micros/notifications/services"
)

// DeleteNotificationHandle handle delete a Notification
func DeleteNotificationHandle(c *fiber.Ctx) error {

	// params from /notifications/id/:notificationId
	notificationId := c.Params("notificationId")
	if notificationId == "" {
		errorMessage := fmt.Sprintf("Notification Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("notificationIdRequired", "Notification id is required!"))
	}

	notificationUUID, uuidErr := uuid.FromString(notificationId)
	if uuidErr != nil {
		errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseUUID", "Can not parse UUID!"))
	}

	// Create service
	notificationService, serviceErr := service.NewNotificationService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("Notification Service Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("notificationService", "Error happened while creating service!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[DeleteNotificationHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	if err := notificationService.DeleteNotificationByOwner(currentUser.UserID, notificationUUID); err != nil {
		errorMessage := fmt.Sprintf("Delete Notification Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("deleteNotification", "Error happened while removing notification!"))

	}

	return c.SendStatus(http.StatusOK)

}

// DeleteNotificationByUserIdHandle handle delete a Notification but userId
func DeleteNotificationByUserIdHandle(c *fiber.Ctx) error {

	// Create service
	notificationService, serviceErr := service.NewNotificationService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("Notification Service Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("notificationService", "Error happened while creating service!"))

	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[DeleteNotificationByUserIdHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	if err := notificationService.DeleteNotificationsByUserId(currentUser.UserID); err != nil {
		errorMessage := fmt.Sprintf("Delete Notification Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("deleteNotification", "Error happened while removing notification!"))
	}

	return c.SendStatus(http.StatusOK)

}
