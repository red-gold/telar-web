package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/red-gold/telar-core/pkg/log"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/profile/database"
	"github.com/red-gold/telar-web/micros/profile/dto"
	service "github.com/red-gold/telar-web/micros/profile/services"
)

// InitProfileIndexHandle handle create a new index
func InitProfileIndexHandle(c *fiber.Ctx) error {

	// Create service
	profileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("Profile service Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	postIndexMap := make(map[string]interface{})
	postIndexMap["fullName"] = "text"
	postIndexMap["objectId"] = 1
	if err := profileService.CreateUserProfileIndex(postIndexMap); err != nil {
		errorMessage := fmt.Sprintf("Create post index Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("createPostIndex", "Error happened while creating post index!"))
	}

	return c.SendStatus(http.StatusOK)

}

// CreateProfileHandle handle create a new profile
func CreateDtoProfileHandle(c *fiber.Ctx) error {

	// Create service
	profileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("Profile service Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	model := new(dto.UserProfile)
	err := c.BodyParser(model)
	if err != nil {
		errorMessage := fmt.Sprintf("parse user profile model %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("parseUserProfileModel", "Error happened while parsing model!"))

	}
	if err = profileService.SaveUserProfile(model); err != nil {
		errorMessage := fmt.Sprintf("Create profile error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("createProfileError", "Error happened while saving user profile!"))
	}

	return c.SendStatus(http.StatusOK)
}
