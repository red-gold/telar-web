package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/micros/profile/database"
	models "github.com/red-gold/telar-web/micros/profile/models"
	service "github.com/red-gold/telar-web/micros/profile/services"
)

type MembersPayload struct {
	Users map[string]interface{} `json:"users"`
}

// ReadDtoProfileHandle a function invocation
func ReadDtoProfileHandle(c *fiber.Ctx) error {

	userId := c.Params("userId")
	log.Info("Read dto profile by userId %s", userId)
	userUUID, uuidErr := uuid.FromString(userId)
	if uuidErr != nil {
		log.Error("Parse UUID %s ", uuidErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseUUIDError", "Can not parse user id!"))
	}
	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		log.Error("NewUserProfileService %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	foundUser, err := userProfileService.FindByUserId(userUUID)
	if err != nil {
		log.Error("FindByUserId %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findByUserId", "Error happened while finding user profile!"))
	}

	if foundUser == nil {
		log.Error("Could not find user " + userUUID.String())
		return c.Status(http.StatusNotFound).JSON(utils.Error("notFoundUser", "Error happened while finding user profile!"))
	}

	return c.JSON(foundUser)

}

// ReadProfileHandle a function invocation
func ReadProfileHandle(c *fiber.Ctx) error {

	userId := c.Params("userId")
	userUUID, uuidErr := uuid.FromString(userId)
	if uuidErr != nil {
		log.Error("Parse UUID %s ", uuidErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseUUIDError", "Can not parse user id!"))
	}

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		log.Error("NewUserProfileService %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	foundUser, err := userProfileService.FindByUserId(userUUID)
	if err != nil {
		log.Error("FindByUserId %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findByUserId", "Error happened while finding user profile!"))
	}

	if foundUser == nil {
		log.Error("Could not find user " + userUUID.String())
		return c.Status(http.StatusNotFound).JSON(utils.Error("notFoundUser", "Error happened while finding user profile!"))
	}

	profileModel := models.MyProfileModel{
		ObjectId:       foundUser.ObjectId,
		FullName:       foundUser.FullName,
		Avatar:         foundUser.Avatar,
		Banner:         foundUser.Banner,
		TagLine:        foundUser.TagLine,
		Birthday:       foundUser.Birthday,
		WebUrl:         foundUser.WebUrl,
		CompanyName:    foundUser.CompanyName,
		FacebookId:     foundUser.FacebookId,
		InstagramId:    foundUser.InstagramId,
		TwitterId:      foundUser.TwitterId,
		AccessUserList: foundUser.AccessUserList,
		Permission:     foundUser.Permission,
	}

	return c.JSON(profileModel)

}

// ReadMyProfileHandle a function invocation to read authed user profile
func ReadMyProfileHandle(c *fiber.Ctx) error {

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		log.Error("NewUserProfileService %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[ReadMyProfileHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}

	foundUser, err := userProfileService.FindByUserId(currentUser.UserID)
	if err != nil {
		log.Error("FindByUserId %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findByUserId", "Error happened while finding user profile!"))
	}
	if foundUser == nil {
		log.Error("Could not find user " + currentUser.UserID.String())
		return c.Status(http.StatusNotFound).JSON(utils.Error("notFoundUser", "Error happened while finding user profile!"))
	}

	profileModel := models.MyProfileModel{
		ObjectId:       foundUser.ObjectId,
		FullName:       foundUser.FullName,
		Avatar:         foundUser.Avatar,
		Banner:         foundUser.Banner,
		TagLine:        foundUser.TagLine,
		Birthday:       foundUser.Birthday,
		WebUrl:         foundUser.WebUrl,
		CompanyName:    foundUser.CompanyName,
		FacebookId:     foundUser.FacebookId,
		InstagramId:    foundUser.InstagramId,
		TwitterId:      foundUser.TwitterId,
		AccessUserList: foundUser.AccessUserList,
		Permission:     foundUser.Permission,
	}
	return c.JSON(profileModel)

}

// DispatchProfilesHandle a function invocation to read authed user profile
func DispatchProfilesHandle(c *fiber.Ctx) error {

	// Parse model object
	model := new(models.DispatchProfilesModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Unmarshal  models.DispatchProfilesModel array %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/dispatchProfilesModelParser", "Error happened while parsing model!"))
	}

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		log.Error("NewUserProfileService %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	foundUsers, err := userProfileService.FindProfileByUserIds(model.UserIds)
	if err != nil {
		log.Error("FindProfileByUserIds %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findProfileByUserIds", "Error happened while finding users profile!"))
	}

	mappedUsers := make(map[string]interface{})
	for _, v := range foundUsers {
		mappedUser := make(map[string]interface{})
		mappedUser["userId"] = v.ObjectId
		mappedUser["fullName"] = v.FullName
		mappedUser["avatar"] = v.Avatar
		mappedUser["banner"] = v.Banner
		mappedUser["tagLine"] = v.TagLine
		mappedUser["lastSeen"] = v.LastSeen
		mappedUser["createdDate"] = v.CreatedDate
		mappedUsers[v.ObjectId.String()] = mappedUser
	}

	actionRoomPayload := &MembersPayload{
		Users: mappedUsers,
	}

	activeRoomAction := Action{
		Type:    "SET_USER_ENTITIES",
		Payload: actionRoomPayload,
	}

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Warn("[DispatchProfilesHandle] Can not get current user")
		currentUser = types.UserContext{}
	}

	userInfoReq := &UserInfoInReq{
		UserId:      currentUser.UserID,
		Username:    currentUser.Username,
		Avatar:      currentUser.Avatar,
		DisplayName: currentUser.DisplayName,
		SystemRole:  currentUser.SystemRole,
	}

	go dispatchAction(activeRoomAction, userInfoReq)

	return c.SendStatus(http.StatusOK)

}

// GetProfileByIds a function invocation to profiles by ids
func GetProfileByIds(c *fiber.Ctx) error {

	// Parse model object
	model := new(models.GetProfilesModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Unmarshal  models.GetProfilesModel array %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/getProfilesModelParser", "Error happened while parsing model!"))
	}

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		log.Error("NewUserProfileService %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	foundUsers, err := userProfileService.FindProfileByUserIds(model.UserIds)
	if err != nil {
		log.Error("FindByUserId %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findByUserId", "Error happened while finding user profile!"))
	}

	return c.JSON(foundUsers)

}
