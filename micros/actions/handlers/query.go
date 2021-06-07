package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/actions/database"
	models "github.com/red-gold/telar-web/micros/actions/models"
	service "github.com/red-gold/telar-web/micros/actions/services"
)

// GetActionRoomHandle handle get a actionRoom
func GetActionRoomHandle(c *fiber.Ctx) error {

	actionRoomId := c.Params("actionRoomId")
	actionRoomUUID, uuidErr := uuid.FromString(actionRoomId)
	if uuidErr != nil {
		errorMessage := fmt.Sprintf("ActionRoom Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("actionRoomIdRequired", "Action room is required!"))
	}
	// Create service
	actionRoomService, serviceErr := service.NewActionRoomService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("ActionRoom Service Error %s", serviceErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("actionRoomService", "Error happend while creating action room service!"))
	}

	foundActionRoom, err := actionRoomService.FindById(actionRoomUUID)
	if err != nil {
		log.Error("[actionRoomService.FindById] %s - %s ", actionRoomUUID.String(), err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findActionRoom", "Can not find action room!"))
	}

	actionRoomModel := models.ActionRoomModel{
		ObjectId:    foundActionRoom.ObjectId,
		OwnerUserId: foundActionRoom.OwnerUserId,
		PrivateKey:  foundActionRoom.PrivateKey,
		AccessKey:   foundActionRoom.AccessKey,
		Status:      foundActionRoom.Status,
		CreatedDate: foundActionRoom.CreatedDate,
	}

	return c.JSON(actionRoomModel)

}

// GetAccessKeyHandle handle get access key
func GetAccessKeyHandle(c *fiber.Ctx) error {

	// Create service
	actionRoomService, serviceErr := service.NewActionRoomService(database.Db)
	if serviceErr != nil {
		errorMessage := fmt.Sprintf("ActionRoom Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("actionRoomIdRequired", "Action room is required!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[GetAccessKeyHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}
	accessKey, err := actionRoomService.GetAccessKey(currentUser.UserID)
	if err != nil {
		log.Error("[actionRoomService.GetAccessKey] %s - %s", currentUser.UserID, err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("getAccessKey", "Can not get access key!"))
	}

	return c.JSON(fiber.Map{
		"accessKey": accessKey,
	})

}

// VerifyAccessKeyHandle handle verify access key
func VerifyAccessKeyHandle(c *fiber.Ctx) error {

	model := new(models.ActionVerifyModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Parse ActionVerifyModel Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseActionVerifyModel", "Error happend while parsing ActionVerifyModel!"))
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
		log.Error("[VerifyAccessKeyHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	isVerified, err := actionRoomService.VerifyAccessKey(currentUser.UserID, model.AccessKey)
	if err != nil {
		errorMessage := fmt.Sprintf("Verify accecc key Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("actionRoomService", "Error happend while verifying access key!"))
	}

	return c.JSON(fiber.Map{
		"isVerified": isVerified,
	})

}
