package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/notifications/database"
	domain "github.com/red-gold/telar-web/micros/notifications/dto"
	models "github.com/red-gold/telar-web/micros/notifications/models"
	service "github.com/red-gold/telar-web/micros/notifications/services"
)

type NotificationAction struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// CreateNotificationHandle handle create a new notification
func CreateNotificationHandle(c *fiber.Ctx) error {

	// Create the model object
	model := new(models.CreateNotificationModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Parse CreateNotificationModel Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("model", "Error happened while parsing model!"))
	}

	// Create service
	notificationService, serviceErr := service.NewNotificationService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("Create notification notification Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/createService", "Error happened while creating service!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[CreateNotificationHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	newNotification := &domain.Notification{
		ObjectId:             model.ObjectId,
		OwnerUserId:          currentUser.UserID,
		OwnerDisplayName:     currentUser.DisplayName,
		OwnerAvatar:          currentUser.Avatar,
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
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("saveNotification", "Error happened while saving notification!"))
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
		userHeaders["uid"] = []string{currentUser.UserID.String()}
		userHeaders["email"] = []string{currentUser.Username}
		userHeaders["avatar"] = []string{currentUser.Avatar}
		userHeaders["displayName"] = []string{currentUser.DisplayName}
		userHeaders["role"] = []string{currentUser.SystemRole}

		_, actionErr := functionCall(http.MethodPost, notificationActionBytes, actionURL, userHeaders)

		if actionErr != nil {
			errorMessage := fmt.Sprintf("Cannot send action request! error: %s", actionErr.Error())
			fmt.Println(errorMessage)
		}
	}()

	return c.JSON(fiber.Map{
		"objectId": newNotification.ObjectId.String(),
	})

}
