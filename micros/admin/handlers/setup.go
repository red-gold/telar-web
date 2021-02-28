package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	models "github.com/red-gold/telar-web/micros/setting/models"
)

// SetupPageHandler creates a handler for logging in
func SetupHandler() func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {

		// Create admin header for http request
		adminHeaders := make(map[string][]string)
		adminHeaders["uid"] = []string{req.UserID.String()}
		adminHeaders["email"] = []string{req.Username}
		adminHeaders["avatar"] = []string{req.Avatar}
		adminHeaders["displayName"] = []string{req.DisplayName}
		adminHeaders["role"] = []string{req.SystemRole}

		// Send request for setting
		getSettingURL := "/setting"
		adminSetting, getSettingErr := functionCallByHeader(http.MethodGet, []byte(""), getSettingURL, adminHeaders)

		if getSettingErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("getSettingError",
						fmt.Sprintf("Cannot get user setting! error: %s", getSettingErr.Error()))},
				nil
		}

		var settingGroupModelMap map[string][]models.GetSettingGroupItemModel
		unmarshalErr := json.Unmarshal(adminSetting, &settingGroupModelMap)
		if unmarshalErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("unmarshalSettingGroupModelError",
						fmt.Sprintf("Cannot unmarshal setting ! error: %s", unmarshalErr.Error()))},
				nil
		}

		setupStatus := "none"
		for _, setting := range settingGroupModelMap["setup"] {
			if setting.Name == "status" {
				setupStatus = setting.Value
			}
		}
		if setupStatus == "completed" {
			return homePageResponse()
		}
		// Create post index
		postIndexURL := "/posts/index"
		_, postIndexErr := functionCall([]byte(""), postIndexURL)

		if postIndexErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("createPostIndexError",
						fmt.Sprintf("Cannot save user profile! error: %s", postIndexErr.Error()))},
				nil
		}

		// Create profile index
		profileIndexURL := "/profile/index"
		_, profileIndexErr := functionCall([]byte(""), profileIndexURL)

		if profileIndexErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("createProfileIndexError",
						fmt.Sprintf("Cannot save user profile! error: %s", profileIndexErr.Error()))},
				nil
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
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("marshaSettigModelError",
						fmt.Sprintf("Cannot marshal setting model! error: %s", marshalErr.Error()))},
				nil
		}

		// Send request for setting
		settingURL := "/setting"
		_, settingErr := functionCallByHeader(http.MethodPost, settingBytes, settingURL, adminHeaders)

		if settingErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("createSettingError",
						fmt.Sprintf("Cannot save user setting! error: %s", settingErr.Error()))},
				nil
		}

		return homePageResponse()
	}
}

// SetupPageHandler creates a handler for logging in
func SetupPageHandler(server.Request) (handler.Response, error) {
	return setupPageResponse()
}

// setupPageResponse login page response template
func setupPageResponse() (handler.Response, error) {
	prettyURL := utils.GetPrettyURLf("/admin/setup")
	html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/start.html", struct {
		SetupAction string
	}{
		SetupAction: prettyURL,
	})
	if parseErr != nil {
		fmt.Printf("Can not parse the html page! error: %s ", parseErr)
		return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("parseHtmlError", "Can not parse the html page!")},
			nil
	}

	return handler.Response{
		Body:       html,
		StatusCode: http.StatusOK,
		Header: map[string][]string{
			"Content-Type": {" text/html; charset=utf-8"},
		},
	}, nil
}

// homePageResponse login page response template
func homePageResponse() (handler.Response, error) {
	html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/home.html", struct {
	}{})
	if parseErr != nil {
		fmt.Printf("Can not parse the html page! error: %s ", parseErr)
		return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("parseHtmlError", "Can not parse the html page!")},
			nil
	}

	return handler.Response{
		Body:       html,
		StatusCode: http.StatusOK,
		Header: map[string][]string{
			"Content-Type": {" text/html; charset=utf-8"},
		},
	}, nil
}
