package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	coreConfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/log"
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
// @Summary Display login page
// @Description Render the login page for Telar Social
// @Tags auth
// @Produce html
// @Success 200 {string} string "OK"
// @Router /auth/login [get]
func LoginPageHandler(c *fiber.Ctx) error {

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
	return loginPageResponse(c, loginData)
}

// LoginAdminHandler creates a handler for logging in telar social
// @Summary Admin login
// @Description Handle admin login for Telar Social
// @Tags auth
// @Accept application/x-www-form-urlencoded
// @Produce html
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.Error "Bad request"
// @Failure 500 {object} utils.Error "Internal server error"
// @Router /auth/login/admin [post]
func LoginAdminHandler(c *fiber.Ctx) error {

	coreConfig := &coreConfig.AppConfig
	adminConfig := ac.AdminConfig

	loginData := &loginPageData{
		title:         "Login - " + *coreConfig.AppName,
		orgName:       *coreConfig.OrgName,
		orgAvatar:     *coreConfig.OrgAvatar,
		appName:       *coreConfig.AppName,
		actionForm:    "",
		resetPassLink: "",
		signupLink:    "",
		message:       "",
	}

	model := &models.LoginModel{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	}

	if model.Username == "" {
		log.Error(" Username is empty")
		loginData.message = "Username is required!"
		return loginPageResponse(c, loginData)
	}

	if model.Password == "" {
		log.Error(" Password is empty")
		loginData.message = "Password is required!"
		return loginPageResponse(c, loginData)
	}
	adminExist, adminCheckErr := checkSetupEnabled()
	if adminCheckErr != nil {
		if adminCheckErr != nil {
			log.Error("Check setup enabled %s", adminCheckErr.Error())
		}
		loginData.message = "Internal error while checking setup!"
		return loginPageResponse(c, loginData)
	}

	var token *string
	log.Info("Admin exist: %t", adminExist)
	if !adminExist {
		adminToken, adminSignupErr := signupAdmin()
		if adminSignupErr != nil {
			if adminSignupErr != nil {
				log.Error("Admin signup error %s", adminSignupErr.Error())
			}
			loginData.message = "Internal error while setup admin!"
			return loginPageResponse(c, loginData)

		}
		token = &adminToken
	} else {
		adminToken, adminLoginErr := loginAdmin(model)
		if adminLoginErr != nil {
			if adminLoginErr != nil {
				log.Error("Admin login error %s", adminLoginErr.Error())
			}
			loginData.message = "Admin login error!"
			return loginPageResponse(c, loginData)

		}
		token = &adminToken
	}
	writeSessionOnCookie(c, *token, &adminConfig)
	prettyURL := utils.GetPrettyURLf("/admin/setup")

	return c.Render("redirect", fiber.Map{
		"URL": prettyURL,
	})

}

// checkSetupEnabled check whether setup is done already
// @Summary Check admin setup
// @Description Check if the admin setup is already done
// @Tags auth
// @Produce json
// @Success 200 {object} AdminCheck "Admin check result"
// @Failure 500 {object} utils.Error "Internal server error"
// @Router /auth/check/admin [post]
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
// @Summary Signup admin
// @Description Signup a new admin for Telar Social
// @Tags auth
// @Produce json
// @Success 200 {object} AdminToken "Admin token"
// @Failure 500 {object} utils.Error "Internal server error"
// @Router /auth/signup/admin [post]
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
// @Summary Admin login
// @Description Login an admin for Telar Social
// @Tags auth
// @Accept json
// @Produce json
// @Param login body models.LoginModel true "Login model"
// @Success 200 {object} AdminToken "Admin token"
// @Failure 500 {object} utils.Error "Internal server error"
// @Router /auth/login/admin [post]
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
// @Summary Render login page
// @Description Render the login page with the provided data
// @Tags auth
// @Produce html
// @Param data body loginPageData true "Login page data"
// @Success 200 {string} string "OK"
// @Router /auth/login/response [post]
func loginPageResponse(c *fiber.Ctx, data *loginPageData) error {
	return c.Render("login", fiber.Map{
		"Title":         data.title,
		"OrgName":       data.orgName,
		"OrgAvatar":     data.orgAvatar,
		"AppName":       data.appName,
		"ActionForm":    data.actionForm,
		"ResetPassLink": data.resetPassLink,
		"SignupLink":    data.signupLink,
		"Message":       data.message,
	})
}
