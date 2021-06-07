package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/setting/database"
	service "github.com/red-gold/telar-web/micros/setting/services"
)

// DeleteUserSettingHandle handle delete a userSetting
func DeleteUserSettingHandle(c *fiber.Ctx) error {

	// params from /userSettings/:userSettingId
	userSettingId := c.Params("userSettingId")
	if userSettingId == "" {
		errorMessage := fmt.Sprintf("UserSetting Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userSettingIdRequired", "userSettingId is Required!"))
	}

	userSettingUUID, uuidErr := uuid.FromString(userSettingId)
	if uuidErr != nil {
		errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("canNotParseUUID", "Can not parse UUID!"))
	}

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("UserSetting Service Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))

	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[DeleteUserSettingHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	if err := userSettingService.DeleteUserSettingByOwner(currentUser.UserID, userSettingUUID); err != nil {
		errorMessage := fmt.Sprintf("Delete UserSetting Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("deleteUserSetting", "Error happened while removing UserSetting!"))

	}
	return c.SendStatus(http.StatusOK)

}

// DeleteUserAllSettingHandle handle delete all userSetting
func DeleteUserAllSettingHandle(c *fiber.Ctx) error {

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("UserSetting Service Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))

	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[DeleteUserAllSettingHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	if err := userSettingService.DeleteUserSettingByOwnerUserId(currentUser.UserID); err != nil {
		errorMessage := fmt.Sprintf("Delete UserSetting Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("deleteUserSetting", "Error happened while removing UserSetting!"))

	}
	return c.SendStatus(http.StatusOK)
}
