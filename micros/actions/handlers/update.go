package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/actions/database"
	domain "github.com/red-gold/telar-web/micros/actions/dto"
	models "github.com/red-gold/telar-web/micros/actions/models"
	service "github.com/red-gold/telar-web/micros/actions/services"
)

// UpdateActionRoomHandle handles updating an existing actionRoom
// @Summary Update an actionRoom
// @Description Updates an existing actionRoom with new data
// @Tags actions
// @Accept json
// @Produce json
// @Security JWT
// @Param Authorization header string true "Authentication" default(Bearer <Add_token_here>)
// @Param body body models.ActionRoomModel true "ActionRoom Model"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.TelarError "Bad request"
// @Failure 500 {object} utils.TelarError "Internal server error"
// @Router /room [put]
func UpdateActionRoomHandle(c *fiber.Ctx) error {

	// Create the model object
	model := new(models.ActionRoomModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Parse ActionRoomModel Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseActionRoomModel", "Error happened while parsing model!"))
	}

	// Create service
	actionRoomService, serviceErr := service.NewActionRoomService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("actionRoom Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("actionRoomService", "Can not create action room service!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[UpdateActionRoomHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	updatedActionRoom := &domain.ActionRoom{
		ObjectId:    model.ObjectId,
		OwnerUserId: currentUser.UserID,
		PrivateKey:  model.PrivateKey,
		AccessKey:   model.AccessKey,
		Status:      model.Status,
		CreatedDate: model.CreatedDate,
	}

	if err := actionRoomService.UpdateActionRoomById(updatedActionRoom); err != nil {
		errorMessage := fmt.Sprintf("Update ActionRoom Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("updateActionRoom", "Can not update action room!"))
	}

	return c.SendStatus(http.StatusOK)

}

// SetAccessKeyHandle handles setting a new access key for the current user
// @Summary Set access key
// @Description Sets a new access key for the current user
// @Tags actions
// @Produce json
// @Security JWT
// @Param Authorization header string true "Authentication" default(Bearer <Add_token_here>)
// @Success 200 {object} object{accessKey=string} "Access key set successfully"
// @Failure 400 {object} utils.TelarError "Bad request"
// @Failure 500 {object} utils.TelarError "Internal server error"
// @Router /room/access-key [put]
func SetAccessKeyHandle(c *fiber.Ctx) error {

	// Create service
	actionRoomService, serviceErr := service.NewActionRoomService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("actionRoom Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("actionRoomService", "Can not create action room service!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[SetAccessKeyHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	accessKey, err := actionRoomService.SetAccessKey(currentUser.UserID)
	if err != nil {
		errorMessage := fmt.Sprintf("Set access key Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("setAccessKey", "Can not set access key!"))
	}

	return c.JSON(fiber.Map{
		"accessKey": accessKey,
	})

}
