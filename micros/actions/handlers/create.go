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

// CreateActionRoomHandle handle create a new actionRoom
func CreateActionRoomHandle(c *fiber.Ctx) error {

	// Create the model object
	model := new(models.CreateActionRoomModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Parse CreateActionRoomModel Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseCreateActionRoomModel", "Error happened while parsing model!"))
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
		log.Error("[CreateActionRoomHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	newActionRoom := &domain.ActionRoom{
		ObjectId:    model.ObjectId,
		OwnerUserId: currentUser.UserID,
		PrivateKey:  model.PrivateKey,
		AccessKey:   model.AccessKey,
		Status:      model.Status,
		CreatedDate: model.CreatedDate,
	}

	if err := actionRoomService.SaveActionRoom(newActionRoom); err != nil {
		errorMessage := fmt.Sprintf("Save ActionRoom Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("saveActionRoom", "Save ActionRoom Error!"))
	}

	return c.JSON(fiber.Map{
		"objectId": newActionRoom.ObjectId.String(),
	})

}
