package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/setting/database"
	domain "github.com/red-gold/telar-web/micros/setting/dto"
	models "github.com/red-gold/telar-web/micros/setting/models"
	service "github.com/red-gold/telar-web/micros/setting/services"
)

// CreateUserSettingHandle handle create a new userSetting
func CreateUserSettingHandle(c *fiber.Ctx) error {

	// Create the model object
	model := new(models.CreateUserSettingModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Unmarshal CreateUserSettingModel Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("parseModel", "Error happened while parsing model!"))

	}

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("userSetting Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[CreateUserSettingHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	newUserSetting := &domain.UserSetting{
		OwnerUserId: currentUser.UserID,
		Name:        model.Name,
		Value:       model.Value,
		Type:        model.Type,
		IsSystem:    false,
	}

	if err := userSettingService.SaveUserSetting(newUserSetting); err != nil {
		errorMessage := fmt.Sprintf("Save UserSetting Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("saveUserSetting", "Error happened while saving UserSetting!"))
	}
	return c.JSON(fiber.Map{
		"objectId": newUserSetting.ObjectId.String(),
	})

}

// CreateSettingGroupHandle handle create a new userSetting
func CreateSettingGroupHandle(c *fiber.Ctx) error {

	// Create the model object
	model := new(models.CreateMultipleSettingsModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Unmarshal CreateSettingGroupModel Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("parseModel", "Error happened while parsing model!"))
	}

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("userSetting Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[CreateSettingGroupHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	var userSettingList []domain.UserSetting
	for _, settings := range model.List {
		for _, setting := range settings.List {
			newUserSetting := domain.UserSetting{
				OwnerUserId: currentUser.UserID,
				Name:        setting.Name,
				Value:       setting.Value,
				Type:        settings.Type,
				IsSystem:    false,
			}
			userSettingList = append(userSettingList, newUserSetting)
		}
	}

	if err := userSettingService.SaveManyUserSetting(userSettingList); err != nil {
		errorMessage := fmt.Sprintf("Save UserSetting Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("saveUserSetting", "Error happened while saving UserSetting!"))
	}

	return c.SendStatus(http.StatusOK)

}
