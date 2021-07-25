package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/profile/database"
	models "github.com/red-gold/telar-web/micros/profile/models"
	service "github.com/red-gold/telar-web/micros/profile/services"
)

// UpdateProfileHandle a function invocation
func UpdateProfileHandle(c *fiber.Ctx) error {

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		log.Error("NewUserProfileService %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	model, unmarshalErr := getUpdateModel(c)
	if unmarshalErr != nil {
		errorMessage := fmt.Sprintf("Error while un-marshaling ProfileUpdateModel: %s",
			unmarshalErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/parseProfileUpdateModel", "Error while parsing body"))

	}
	log.Info("model %v", model)

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[UpdateProfileHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	log.Info("Update profile %s - %v", currentUser.UserID, model)
	err := userProfileService.UpdateUserProfileById(currentUser.UserID, model)
	if err != nil {
		log.Error("Could not update user profile! %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Could not update user profile!"))
	}

	return c.SendStatus(http.StatusOK)

}

// UpdateLastSeen a function invocation
func UpdateLastSeen(c *fiber.Ctx) error {

	model := new(models.UpdateLastSeenModel)

	unmarshalErr := c.BodyParser(model)
	if unmarshalErr != nil {
		errorMessage := fmt.Sprintf("Unmarshal models.UpdateLastSeenModel %s",
			unmarshalErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("bodyParserUpdateLastSeenModel", "Could not parse UpdateLastSeenModel!"))
	}

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("internal/userProfileService", "Internal error happened while creating userProfileService!"))
	}

	err := userProfileService.UpdateLastSeenNow(model.UserId)
	if err != nil {
		errorMessage := fmt.Sprintf("Update last seen %s",
			err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("updateLastSeen", "Error happened while updating last seen"))

	}

	return c.SendStatus(http.StatusOK)

}

// IncreaseFollowCount a function invocation
func IncreaseFollowCount(c *fiber.Ctx) error {

	// params from /follow/inc/:inc/:userId
	userId := c.Params("userId")
	if userId == "" {
		errorMessage := fmt.Sprintf("User Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userIdRequired", errorMessage))
	}

	userUUID, uuidErr := uuid.FromString(userId)
	if uuidErr != nil {
		errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userIdIsNotValid", "Post id is not valid!"))
	}

	incParam := c.Params("inc")
	inc, err := strconv.Atoi(incParam)
	if err != nil {
		log.Error("Wrong inc param %s - %s", incParam, err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidIncParam", "Wrong inc param!"))

	}

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("internal/userProfileService", "Internal error happened while creating userProfileService!"))
	}

	err = userProfileService.IncreaseFollowCount(userUUID, inc)
	if err != nil {
		errorMessage := fmt.Sprintf("Update follow count %s",
			err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("updateFollowCount", "Error happened while updating follow count!"))

	}

	return c.SendStatus(http.StatusOK)

}

// IncreaseFollowerCount a function invocation
func IncreaseFollowerCount(c *fiber.Ctx) error {

	// params from /follower/inc/:inc/:userId
	userId := c.Params("userId")
	if userId == "" {
		errorMessage := fmt.Sprintf("User Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userIdRequired", errorMessage))
	}

	userUUID, uuidErr := uuid.FromString(userId)
	if uuidErr != nil {
		errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userIdIsNotValid", "Post id is not valid!"))
	}

	incParam := c.Params("inc")
	inc, err := strconv.Atoi(incParam)
	if err != nil {
		log.Error("Wrong inc param %s - %s", incParam, err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidIncParam", "Wrong inc param!"))

	}

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("internal/userProfileService", "Internal error happened while creating userProfileService!"))
	}

	err = userProfileService.IncreaseFollowerCount(userUUID, inc)
	if err != nil {
		errorMessage := fmt.Sprintf("Update follower count %s",
			err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("updateFollowerCount", "Error happened while updating follower count!"))

	}

	return c.SendStatus(http.StatusOK)

}
