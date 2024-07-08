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

// @Summary Create a new index
// @Description Create a new index for user profiles
// @Tags profiles
// @Accept  json
// @Produce  json
// @Security JWT
// @Param Authorization header string true "Authentication" default(Bearer <Add_token_here>)
// @Success 200
// @Failure 400 {object} utils.TelarError
// @Failure 500 {object} utils.TelarError
// @Router /index [post]
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

// @Summary Create a new profile
// @Description Create a new user profile
// @Tags profiles
// @Accept  json
// @Produce  json
// @Security JWT
// @Param Authorization header string true "Authentication" default(Bearer <Add_token_here>)
// @Param   body  body     dto.UserProfile  true "User profile model"
// @Success 200
// @Failure 400 {object} utils.TelarError
// @Failure 500 {object} utils.TelarError
// @Router /dto [post]
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
