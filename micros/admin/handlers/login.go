package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	handler "github.com/openfaas-incubator/go-function-sdk"
	coreConfig "github.com/red-gold/telar-core/config"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	ac "github.com/red-gold/telar-web/micros/admin/config"
	models "github.com/red-gold/telar-web/micros/auth/models"
)

// Login page data template
type loginPageData struct {
	title         string
	orgName       string
	orgAvatar     string
	appName       string
	actionForm    string
	resetPassLink string
	signupLink    string
	message       string
}

// Admin check
type AdminCheck struct {
	Success bool `json:"success"`
	Admin   bool `json:"admin"`
}

// Admin token
type AdminToken struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

// LoginPageHandler creates a handler for logging in
func LoginPageHandler(server.Request) (handler.Response, error) {

	appConfig := coreConfig.AppConfig
	prettyURL := utils.GetPrettyURLf("/auth")
	loginData := &loginPageData{
		title:         "Login - Telar Social",
		orgName:       *appConfig.OrgName,
		orgAvatar:     *appConfig.OrgAvatar,
		appName:       *appConfig.AppName,
		actionForm:    "",
		resetPassLink: "",
		signupLink:    prettyURL + "/signup",
		message:       "",
	}
	return loginPageResponse(loginData)
}

// LoginAdminHandler creates a handler for logging in telar social
func LoginAdminHandler(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		coreConfig := &coreConfig.AppConfig
		adminConfig := ac.AdminConfig

		loginData := &loginPageData{
			title:         "Login - Telar Social",
			orgName:       *coreConfig.OrgName,
			orgAvatar:     *coreConfig.OrgAvatar,
			appName:       *coreConfig.AppName,
			actionForm:    "",
			resetPassLink: "",
			signupLink:    "",
			message:       "",
		}

		var query *url.Values
		if len(req.Body) > 0 {
			q, parseErr := url.ParseQuery(string(req.Body))
			if parseErr != nil {
				errorMessage := fmt.Sprintf("{error: 'parse SignupTokenModel (%s): %s'}",
					req.Body, parseErr.Error())
				return handler.Response{StatusCode: http.StatusBadRequest, Body: []byte(errorMessage)},
					parseErr

			}
			query = &q

		}

		model := &models.LoginModel{
			Username: query.Get("username"),
			Password: query.Get("password"),
		}

		if model.Username == "" {
			fmt.Printf("\n Username is empty\n")
			loginData.message = "Username is required!"
			return loginPageResponse(loginData)
		}

		if model.Password == "" {
			fmt.Printf("\n Password is empty\n")
			loginData.message = "Password is required!"
			return loginPageResponse(loginData)
		}

		adminExist, adminCheckErr := checkSetupEnabled()
		if adminCheckErr != nil {
			errorMessage := fmt.Sprintf("Admin check error: %s", adminCheckErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("adminCheckError", errorMessage)}, nil
		}
		var token *string
		fmt.Printf("Admin exist: %t", adminExist)
		if !adminExist {
			adminToken, adminSignupErr := signupAdmin()
			if adminSignupErr != nil {
				errorMessage := fmt.Sprintf("Admin signup error: %s", adminSignupErr.Error())
				return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("adminSignupError", errorMessage)}, nil
			}
			token = &adminToken
		} else {
			adminToken, adminLoginErr := loginAdmin(model)
			if adminLoginErr != nil {
				errorMessage := fmt.Sprintf("Admin login error: %s", adminLoginErr.Error())
				return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("adminLoginError", errorMessage)}, nil
			}
			token = &adminToken
		}
		writeSessionOnCookie(w, *token, &adminConfig)
		prettyURL := utils.GetPrettyURLf("/admin/setup")

		http.Redirect(w, r, prettyURL, http.StatusTemporaryRedirect)

		return handler.Response{
			StatusCode: http.StatusOK,
		}, nil
	}
}

// checkSetupEnabled check whether setup is done already
func checkSetupEnabled() (bool, error) {
	url := "/auth/check/admin"
	resData, functionCallErr := functionCall([]byte(""), url, http.MethodPost)
	if functionCallErr != nil {
		return false, functionCallErr
	}

	var adminCheck AdminCheck
	jsonErr := json.Unmarshal(resData, &adminCheck)
	if jsonErr != nil {
		return false, fmt.Errorf("failed to unmarshal admin check json, error: %s", jsonErr.Error())
	}
	return adminCheck.Admin, nil
}

// signupAdmin signup admin
func signupAdmin() (string, error) {
	url := "/auth/signup/admin"
	resData, functionCallErr := functionCall([]byte(""), url, http.MethodPost)
	if functionCallErr != nil {
		return "", functionCallErr
	}
	var adminsignup AdminToken
	jsonErr := json.Unmarshal(resData, &adminsignup)
	if jsonErr != nil {
		return "", fmt.Errorf("failed to unmarshal admin check json, error: %s", jsonErr.Error())
	}
	return adminsignup.Token, nil
}

// loginAdmin login admin
func loginAdmin(model *models.LoginModel) (string, error) {
	url := "/auth/login/admin"
	bytesOut, _ := json.Marshal(model)
	resData, functionCallErr := functionCall(bytesOut, url, http.MethodPost)
	if functionCallErr != nil {
		return "", functionCallErr
	}
	var adminsignup AdminToken
	jsonErr := json.Unmarshal(resData, &adminsignup)
	if jsonErr != nil {
		return "", fmt.Errorf("failed to unmarshal admin check json, error: %s", jsonErr.Error())
	}
	return adminsignup.Token, nil
}

// loginPageResponse login page response template
func loginPageResponse(data *loginPageData) (handler.Response, error) {
	html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/login.html", struct {
		Title         string
		OrgName       string
		OrgAvatar     string
		AppName       string
		ActionForm    string
		ResetPassLink string
		SignupLink    string
		Message       string
	}{
		Title:         data.title,
		OrgName:       data.orgName,
		OrgAvatar:     data.orgAvatar,
		AppName:       data.appName,
		ActionForm:    data.actionForm,
		ResetPassLink: data.resetPassLink,
		SignupLink:    data.signupLink,
		Message:       data.message,
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
