package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	utils "github.com/red-gold/telar-core/utils"
	cf "github.com/red-gold/telar-web/micros/auth/config"
	models "github.com/red-gold/telar-web/micros/auth/models"
	"github.com/red-gold/telar-web/micros/auth/provider"
)

// UpdateProfileHandle a function invocation
func UpdateProfileHandle(c *fiber.Ctx) error {

	authConfig := &cf.AuthConfig

	model := new(models.ProfileUpdateModel)
	unmarshalErr := c.BodyParser(model)
	if unmarshalErr != nil {
		errorMessage := fmt.Sprintf("Error while un-marshaling ProfileUpdateModel: %s",
			unmarshalErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/marshalProfileUpdateModelError", "Can not parse ProfileUpdateModel!"))

	}

	currentUser, ok := c.Locals(types.UserCtxName).(types.UserContext)
	if !ok {
		log.Error("[UpdateProfileHandle] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}
	err := updateUserProfile(model, currentUser.UserID, currentUser.Username, currentUser.Avatar, currentUser.DisplayName, currentUser.SystemRole)
	if err != nil {
		log.Error("Can not update user profile %s ", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/updateUserProfile", "Can not update user profile!"))

	}

	tokenModel := &TokenModel{
		token:            ProviderAccessToken{},
		oauthProvider:    nil,
		providerName:     "telar",
		profile:          &provider.Profile{Name: model.FullName, ID: currentUser.UserID.String(), Login: currentUser.Username},
		organizationList: *config.AppConfig.OrgName,
		claim: UserClaim{
			DisplayName: model.FullName,
			Email:       currentUser.Username,
			Avatar:      model.Avatar,
			UserId:      currentUser.UserID.String()},
	}
	session, err := createToken(tokenModel)
	if err != nil {
		log.Error("Error creating session: %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/createToken", "Internal server error creating token!"))
	}

	// Write session on cookie
	writeSessionOnCookie(c, session, authConfig)
	log.Info("\nSession is created: %s \n", session)
	return c.SendStatus(http.StatusOK)

}
