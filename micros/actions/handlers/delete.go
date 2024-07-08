package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/actions/database"
	service "github.com/red-gold/telar-web/micros/actions/services"
)

// DeleteActionRoomHandle handles the deletion of an actionRoom
// @Summary Delete an actionRoom
// @Description Handles the deletion of an actionRoom by its ID
// @Tags actions
// @Produce json
// @Security JWT
// @Param Authorization header string true "Authentication" default(Bearer <Add_token_here>)
// @Param roomId path string true "ActionRoom ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.TelarError "Bad request"
// @Failure 500 {object} utils.TelarError "Internal server error"
// @Router /room/{roomId} [delete]
func DeleteActionRoomHandle(c *fiber.Ctx) error {

	// params from /actions/room/:roomId
	actionRoomId := c.Params("roomId")
	if actionRoomId == "" {
		errorMessage := fmt.Sprintf("ActionRoom Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("actionRoomIdRequired", "Action room is required!"))
	}

	actionRoomUUID, uuidErr := uuid.FromString(actionRoomId)
	if uuidErr != nil {
		errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("uuidError", "Can not parse uuid!"))
	}

	// Create service
	actionRoomService, serviceErr := service.NewActionRoomService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("ActionRoom Service Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("actionRoomService", "Error happend while creating action room service!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[DeleteActionRoomHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	if err := actionRoomService.DeleteActionRoomByOwner(currentUser.UserID, actionRoomUUID); err != nil {
		errorMessage := fmt.Sprintf("Delete ActionRoom Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("actionRoomService", "Error happend while removing action room!"))

	}

	return c.SendStatus(http.StatusOK)

}
