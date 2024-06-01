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

// @Summary Read DTO profile by user ID
// @Description Read DTO profile by user ID
// @Tags profiles
// @Accept  json
// @Produce  json
// @Param   userId  path     string  true "User ID"
// @Success 200 {object} UserProfile
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /profiles/{userId} [get]
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

	foundUserChan, errChan := userProfileService.FindByUserId(userUUID)
	foundUser, err := <-foundUserChan, <-errChan
	if err != nil {
		log.Error("FindByUserId %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findByUserId", "Error happened while finding user profile!"))
	}

	if foundUser == nil {
		log.Error("[ReadDtoProfileHandle] Could not find user " + userUUID.String())
		return c.Status(http.StatusNotFound).JSON(utils.Error("notFoundUser", "Error happened while finding user profile!"))
	}

	return c.JSON(foundUser)

}

// @Summary Read profile by user ID
// @Description Read profile by user ID
// @Tags profiles
// @Accept  json
// @Produce  json
// @Param   userId  path     string  true "User ID"
// @Success 200 {object} MyProfileModel
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /profiles/{userId} [get]
func ReadProfileHandle(c *fiber.Ctx) error {

	userId := c.Params("userId")
	if userId == "" {
		errorMessage := fmt.Sprintf("User Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userIdRequired", errorMessage))
	}
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

	foundUserChan, errChan := userProfileService.FindByUserId(userUUID)
	foundUser, err := <-foundUserChan, <-errChan
	if err != nil {
		log.Error("FindByUserId %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findByUserId", "Error happened while finding user profile!"))
	}

	if foundUser == nil {
		log.Error("[ReadProfileHandle] Could not find user " + userUUID.String())
		return c.Status(http.StatusNotFound).JSON(utils.Error("notFoundUser", "Error happened while finding user profile!"))
	}

	profileModel := models.MyProfileModel{
		ObjectId:       foundUser.ObjectId,
		FullName:       foundUser.FullName,
		SocialName:     foundUser.SocialName,
		Avatar:         foundUser.Avatar,
		Banner:         foundUser.Banner,
		TagLine:        foundUser.TagLine,
		Birthday:       foundUser.Birthday,
		Address:        foundUser.Address,
		LastSeen:       foundUser.LastSeen,
		FollowCount:    foundUser.FollowCount,
		FollowerCount:  foundUser.FollowerCount,
		WebUrl:         foundUser.WebUrl,
		CompanyName:    foundUser.CompanyName,
		FacebookId:     foundUser.FacebookId,
		InstagramId:    foundUser.InstagramId,
		TwitterId:      foundUser.TwitterId,
		LinkedInId:     foundUser.LinkedInId,
		AccessUserList: foundUser.AccessUserList,
		Permission:     foundUser.Permission,
		CreatedDate:    foundUser.CreatedDate,
	}

	return c.JSON(profileModel)

}

// @Summary Get user profile by social name
// @Description Get user profile by social name
// @Tags profiles
// @Accept  json
// @Produce  json
// @Param   name  path     string  true "Social name"
// @Success 200 {object} MyProfileModel
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /profiles/{name} [get]
func GetBySocialName(c *fiber.Ctx) error {

	socialName := c.Params("name")
	if socialName == "" {
		errorMessage := fmt.Sprintf("Social name is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("socialNameRequired", errorMessage))
	}

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		log.Error("NewUserProfileService %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	foundUserChan, errChan := userProfileService.FindBySocialName(socialName)
	foundUser, err := <-foundUserChan, <-errChan
	if err != nil {
		log.Error("findBySocialName %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findBySocialName", "Error happened while finding user profile!"))
	}

	if foundUser == nil {
		log.Error("[GetBySocialName] Could not find user " + socialName)
		return c.Status(http.StatusNotFound).JSON(utils.Error("notFoundUser", "Error happened while finding user profile!"))
	}

	profileModel := models.MyProfileModel{
		ObjectId:       foundUser.ObjectId,
		FullName:       foundUser.FullName,
		SocialName:     foundUser.SocialName,
		Avatar:         foundUser.Avatar,
		Banner:         foundUser.Banner,
		TagLine:        foundUser.TagLine,
		Birthday:       foundUser.Birthday,
		Address:        foundUser.Address,
		LastSeen:       foundUser.LastSeen,
		FollowCount:    foundUser.FollowCount,
		FollowerCount:  foundUser.FollowerCount,
		WebUrl:         foundUser.WebUrl,
		CompanyName:    foundUser.CompanyName,
		FacebookId:     foundUser.FacebookId,
		InstagramId:    foundUser.InstagramId,
		TwitterId:      foundUser.TwitterId,
		LinkedInId:     foundUser.LinkedInId,
		AccessUserList: foundUser.AccessUserList,
		Permission:     foundUser.Permission,
		CreatedDate:    foundUser.CreatedDate,
	}

	return c.JSON(profileModel)

}

// @Summary Read my profile
// @Description Read my profile
// @Tags profiles
// @Accept  json
// @Produce  json
// @Success 200 {object} MyProfileModel
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /profiles/me [get]
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

	foundUserChan, errUserChan := userProfileService.FindByUserId(currentUser.UserID)

	foundUser, errUser, actionAccessKey := <-foundUserChan, <-errUserChan, <-GetActionAccessKeyAsync(getUserInfoReq(c))
	if errUser != nil {
		log.Error("FindByUserId %s", errUser.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findByUserId", "Error happened while finding user profile!"))
	}
	if actionAccessKey.Error != nil {
		log.Error("GetActionAccessKeyAsync %s", actionAccessKey.Error.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/getActionAccessKeyAsync", "Error happened while getting action access key!"))
	}
	if foundUser == nil {
		log.Error("[ReadMyProfileHandle] Could not find user " + currentUser.UserID.String())
		return c.Status(http.StatusNotFound).JSON(utils.Error("notFoundUser", "Error happened while finding user profile!"))
	}

	profileModel := models.MyProfileModel{
		ObjectId:       foundUser.ObjectId,
		FullName:       foundUser.FullName,
		SocialName:     foundUser.SocialName,
		Avatar:         foundUser.Avatar,
		Banner:         foundUser.Banner,
		TagLine:        foundUser.TagLine,
		Birthday:       foundUser.Birthday,
		CompanyName:    foundUser.CompanyName,
		Country:        foundUser.Country,
		Address:        foundUser.Address,
		LastSeen:       foundUser.LastSeen,
		Phone:          foundUser.Phone,
		WebUrl:         foundUser.WebUrl,
		FollowCount:    foundUser.FollowCount,
		FollowerCount:  foundUser.FollowerCount,
		FacebookId:     foundUser.FacebookId,
		InstagramId:    foundUser.InstagramId,
		TwitterId:      foundUser.TwitterId,
		LinkedInId:     foundUser.LinkedInId,
		AccessUserList: foundUser.AccessUserList,
		Permission:     foundUser.Permission,
	}
	c.Set("action-access-key", actionAccessKey.AccessKey)
	return c.JSON(profileModel)

}

// @Summary Dispatch profiles
// @Description Dispatch profiles
// @Tags profiles
// @Accept  json
// @Produce  json
// @Param   request  body     models.DispatchProfilesModel  true "Dispatch profiles model"
// @Success 200
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /profiles/dispatch [post]
func DispatchProfilesHandle(c *fiber.Ctx) error {

	// Parse model object
	model := new(models.DispatchProfilesModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Unmarshal  models.DispatchProfilesModel array %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/dispatchProfilesModelParser", "Error happened while parsing model!"))
	}

	if len(model.UserIds) == 0 {
		errorMessage := "UserIds is required!"
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userIdsRequired", errorMessage))
	}

	// Create service
	userProfileService, serviceErr := service.NewUserProfileService(database.Db)
	if serviceErr != nil {
		log.Error("NewUserProfileService %s", serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userProfileService", "Error happened while creating userProfileService!"))
	}

	foundUsers, err := userProfileService.FindProfileByUserIds(model.UserIds)
	if err != nil {
		log.Error("[DispatchProfilesHandle] FindProfileByUserIds %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findProfileByUserIds", "Error happened while finding users profile!"))
	}

	mappedUsers := make(map[string]interface{})
	for _, v := range foundUsers {
		mappedUser := make(map[string]interface{})
		mappedUser["userId"] = v.ObjectId
		mappedUser["fullName"] = v.FullName
		mappedUser["socialName"] = v.SocialName
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
	log.Info("Current USER %v", currentUser)

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

// @Summary Get profiles by IDs
// @Description Get profiles by IDs
// @Tags profiles
// @Accept  json
// @Produce  json
// @Param   request  body     models.GetProfilesModel  true "Get profiles model"
// @Success 200 {array} UserProfile
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security BearerAuth
// @Router /profiles [post]
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
