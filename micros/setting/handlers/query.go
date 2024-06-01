package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/pkg/parser"
	"github.com/red-gold/telar-core/types"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/setting/database"
	models "github.com/red-gold/telar-web/micros/setting/models"
	service "github.com/red-gold/telar-web/micros/setting/services"
)

type UserSettingQueryModel struct {
	Search string    `query:"search"`
	Page   int64     `query:"page"`
	Owner  uuid.UUID `query:"owner"`
	Type   int       `query:"type"`
}

// @Summary Query user settings
// @Description Query user settings based on search, owner, and type
// @Tags user-settings
// @Accept  json
// @Produce  json
// @Param   query  body     UserSettingQueryModel  true "Query parameters"
// @Success 200 {array}   models.UserSetting
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /userSettings [get]
func QueryUserSettingHandle(c *fiber.Ctx) error {

	query := new(UserSettingQueryModel)

	if err := parser.QueryParser(c, query); err != nil {
		log.Error("[QueryUserSettingHandle] QueryParser %s", err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("queryParser", "Error happened while parsing query!"))
	}

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		log.Error("[NewUserSettingService] %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))
	}

	userSettingList, err := userSettingService.QueryUserSetting(query.Search, &query.Owner, &query.Type, "created_date", query.Page)
	if err != nil {
		log.Error("[QueryUserSetting] %s ", err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("queryUserSetting", "Can not query user setting!"))
	}

	return c.JSON(userSettingList)

}

// @Summary Get all user settings
// @Description Get all user settings for the current user
// @Tags user-settings
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string][]models.GetSettingGroupItemModel
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /userSettings [get]
func GetAllUserSetting(c *fiber.Ctx) error {

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		log.Error("[NewUserSettingService] %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[GetAllUserSetting] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	userSettingList, err := userSettingService.GetAllUserSetting(currentUser.UserID)
	if err != nil {
		log.Error("[GetAllUserSetting] %s ", err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("getAllUserSetting", "Can not get user settings!"))
	}

	groupSettingsMap := make(map[string][]models.GetSettingGroupItemModel)
	for _, setting := range userSettingList {

		settingModel := models.GetSettingGroupItemModel{
			ObjectId: setting.ObjectId,
			Name:     setting.Name,
			Value:    setting.Value,
			IsSystem: setting.IsSystem,
		}
		groupSettingsMap[setting.Type] = append(groupSettingsMap[setting.Type], settingModel)
	}

	return c.JSON(groupSettingsMap)

}

// @Summary Get all user settings by type
// @Description Get all user settings by type for the current user
// @Tags user-settings
// @Accept  json
// @Produce  json
// @Param   type  path     string true "Setting type"
// @Success 200 {object} models.GetSettingGroupModel
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /userSettings/{type} [get]
func GetAllUserSettingByType(c *fiber.Ctx) error {

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		log.Error("[NewUserSettingService] %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))
	}

	settingType := c.Params("type")
	if settingType == "" {
		errorMessage := fmt.Sprintf("Error setting type can not be empty.")
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("settingTypeRquired", "Error setting type can not be empty.!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[GetAllUserSettingByType] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	userSettingList, err := userSettingService.GetAllUserSettingByType(currentUser.UserID, settingType)
	if err != nil {
		log.Error("[GetAllUserSettingByType] %s ", err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("getAllUserSettingByType", "Can not get user settings by type!"))
	}

	var groupSettingItems []models.GetSettingGroupItemModel
	for _, setting := range userSettingList {

		settingModel := models.GetSettingGroupItemModel{
			ObjectId: setting.ObjectId,
			Name:     setting.Name,
			Value:    setting.Value,
			IsSystem: setting.IsSystem,
		}
		groupSettingItems = append(groupSettingItems, settingModel)
	}
	groupSettingsModel := models.GetSettingGroupModel{
		Type:        settingType,
		CreatedDate: userSettingList[0].CreatedDate,
		OwnerUserId: currentUser.UserID,
		List:        groupSettingItems,
	}

	return c.JSON(groupSettingsModel)

}

// @Summary Get a user setting by ID
// @Description Get a user setting by ID for the current user
// @Tags user-settings
// @Accept  json
// @Produce  json
// @Param   userSettingId  path     string true "User setting ID"
// @Success 200 {object} models.UserSettingModel
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /userSettings/{userSettingId} [get]
func GetUserSettingHandle(c *fiber.Ctx) error {

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		log.Error("[NewUserSettingService] %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))
	}

	userSettingId := c.Params("userSettingId")
	userSettingUUID, uuidErr := uuid.FromString(userSettingId)
	if uuidErr != nil {
		log.Error("Can not parse userSetting id! %s", uuidErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseUUIDError", "Can not parse userSetting id!"))
	}

	foundUserSetting, err := userSettingService.FindById(userSettingUUID)
	if err != nil {
		log.Error("[userSettingService.FindById] %s ", err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findUserSetting", "Can not find user settings by id!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[GetUserSettingHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	userSettingModel := models.UserSettingModel{
		ObjectId:    foundUserSetting.ObjectId,
		OwnerUserId: currentUser.UserID,
		CreatedDate: foundUserSetting.CreatedDate,
		Name:        foundUserSetting.Name,
		Value:       foundUserSetting.Value,
		Type:        foundUserSetting.Type,
		IsSystem:    foundUserSetting.IsSystem,
	}

	return c.JSON(userSettingModel)

}

// @Summary Get settings by user IDs
// @Description Get settings by user IDs
// @Tags user-settings
// @Accept  json
// @Produce  json
// @Param   request  body     models.GetSettingsModel  true "Get settings model"
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /settings [post]
func GetSettingByUserIds(c *fiber.Ctx) error {

	// Parse model object
	model := new(models.GetSettingsModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Unmarshal  models.GetProfilesModel array %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("profilesModelParser",
			"Can not parse model!"))
	}

	if !(len(model.UserIds) > 0) {
		log.Error("model.UserIds is empty ")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userIdRequired",
			"User id is required!"))

	}

	// Create service
	userSettingService, serviceErr := service.NewUserSettingService(database.Db)
	if serviceErr != nil {
		log.Error("[NewUserSettingService] %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userSettingService", "Error happened while creating userSettingService!"))
	}

	foundUserSetting, err := userSettingService.FindSettingByUserIds(model.UserIds, model.Type)
	if err != nil {
		log.Error("[userSettingService.FindSettingByUserIds] %s ", err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findUserSetting", "Can not find users settings by ids!"))
	}

	mappedSetting := make(map[string]string)
	for _, setting := range foundUserSetting {
		key := fmt.Sprintf("%s:%s:%s", setting.OwnerUserId, setting.Type, setting.Name)
		mappedSetting[key] = setting.Value
	}

	return c.JSON(mappedSetting)

}
