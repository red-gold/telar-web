package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	utils "github.com/red-gold/telar-core/utils"
	models "github.com/red-gold/telar-web/micros/setting/models"
)

// SetupPageHandler creates a handler for logging in
func SetupHandler(c *fiber.Ctx) error {

	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		log.Error("[SetupHandler] Can not get current user")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCurrentUser",
			"Can not get current user"))
	}
	// Create admin header for http request
	adminHeaders := make(map[string][]string)
	adminHeaders["uid"] = []string{currentUser.UserID.String()}
	adminHeaders["email"] = []string{currentUser.Username}
	adminHeaders["avatar"] = []string{currentUser.Avatar}
	adminHeaders["displayName"] = []string{currentUser.DisplayName}
	adminHeaders["role"] = []string{currentUser.SystemRole}

	// Send request for setting
	getSettingURL := "/setting"
	adminSetting, getSettingErr := functionCallByHeader(http.MethodGet, []byte(""), getSettingURL, adminHeaders)

	if getSettingErr != nil {
		log.Error("[functionCallByHeader] %s - %s", getSettingURL, getSettingErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/functionCallByHeader", "Error happened while getting settings!"))
	}

	var settingGroupModelMap map[string][]models.GetSettingGroupItemModel
	unmarshalErr := json.Unmarshal(adminSetting, &settingGroupModelMap)
	if unmarshalErr != nil {
		log.Error("[unmarshalSettingGroupModel] %s", unmarshalErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/unmarshalSettingGroupModel", "Error happened while unmarshaling settings!"))
	}

	setupStatus := "none"
	for _, setting := range settingGroupModelMap["setup"] {
		if setting.Name == "status" {
			setupStatus = setting.Value
		}
	}
	if setupStatus == "completed" {
		return homePageResponse(c)
	}
	// Create post index
	postIndexURL := "/posts/index"
	_, postIndexErr := functionCall([]byte(""), postIndexURL, http.MethodPost)

	if postIndexErr != nil {
		log.Error("[createPostIndex] %s - %s", postIndexURL, postIndexErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/createPostIndex", "Error happened while creating post index!"))
	}

	// Create profile index
	profileIndexURL := "/profile/index"
	_, profileIndexErr := functionCall([]byte(""), profileIndexURL, http.MethodPost)

	if profileIndexErr != nil {
		log.Error("[profileIndex] %s - %s", postIndexURL, profileIndexErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/profileIndex", "Error happened while creating profile index!"))
	}

	// Create setting for setup compeleted status
	settingModel := models.CreateSettingGroupModel{
		Type: "setup",
		List: []models.SettingGroupItemModel{
			{
				Name:  "status",
				Value: "completed",
			},
		},
	}

	settingBytes, marshalErr := json.Marshal(&settingModel)
	if marshalErr != nil {
		log.Error("[marshaSettigModel] %s", marshalErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/marshalErr", "Error happened while marshaling settingBytes!"))
	}

	// Send request for setting
	settingURL := "/setting"
	_, settingErr := functionCallByHeader(http.MethodPost, settingBytes, settingURL, adminHeaders)

	if settingErr != nil {
		log.Error("[createSetting] %s - %s", settingURL, marshalErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/createSetting", "Error happened while creating setting!"))
	}

	return homePageResponse(c)

}

// SetupPageHandler creates a handler for logging in
func SetupPageHandler(c *fiber.Ctx) error {
	return setupPageResponse(c)
}

// setupPageResponse login page response template
func setupPageResponse(c *fiber.Ctx) error {
	prettyURL := utils.GetPrettyURLf("/admin/setup")

	return c.Render("start", fiber.Map{
		"SetupAction": prettyURL,
	})
}

// homePageResponse login page response template
func homePageResponse(c *fiber.Ctx) error {
	return c.Render("home", fiber.Map{})
}
