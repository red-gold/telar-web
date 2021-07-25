package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	coreConfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/log"
	utils "github.com/red-gold/telar-core/utils"
	authConfig "github.com/red-gold/telar-web/micros/auth/config"
	"github.com/red-gold/telar-web/micros/auth/database"
	models "github.com/red-gold/telar-web/micros/auth/models"
	"github.com/red-gold/telar-web/micros/auth/provider"
	service "github.com/red-gold/telar-web/micros/auth/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	githubLink    string
	message       string
}

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8000/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

// LoginGithubHandler creates a handler for logging in github
func LoginGithubHandler(c *fiber.Ctx) error {

	config := authConfig.AuthConfig
	log.Info("Login to path %s", c.Path())

	resource := "/"

	if val := c.Query("r"); len(val) > 0 {
		resource = val
	}

	u := buildGitHubURL(&config, resource, "read:org,read:user,user:email")
	return c.Redirect(u.String(), http.StatusTemporaryRedirect)

}

// LoginGoogleHandler makes a handler for OAuth 2.0 redirects
func LoginGoogleHandler(c *fiber.Ctx) error {

	// Create oauthState cookie
	oauthState := generateStateOauthCookie(c)

	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the the state query parameter on your redirect callback.
	*/
	u := googleOauthConfig.AuthCodeURL(oauthState)
	return c.Redirect(u, http.StatusTemporaryRedirect)

}

// LoginPageHandler creates a handler for logging in
func LoginPageHandler(c *fiber.Ctx) error {

	appConfig := coreConfig.AppConfig
	authConfig := &authConfig.AuthConfig
	prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)
	loginData := &loginPageData{
		title:         "Login - Telar Social",
		orgName:       *appConfig.OrgName,
		orgAvatar:     *appConfig.OrgAvatar,
		appName:       *appConfig.AppName,
		actionForm:    "",
		resetPassLink: prettyURL + "/password/forget",
		signupLink:    prettyURL + "/signup",
		githubLink:    prettyURL + "/login/github",
		message:       "",
	}
	return loginPageResponse(c, loginData)
}

// LoginTelarHandler creates a handler for logging in telar social
func LoginTelarHandler(c *fiber.Ctx) error {

	model := &models.LoginModel{
		Username:     c.FormValue("username"),
		Password:     c.FormValue("password"),
		ResponseType: c.FormValue("responseType"),
	}

	if model.ResponseType == SPAResponseType {
		return LoginTelarHandlerSPA(c, model)
	}
	return LoginTelarHandlerSSR(c, model)
}

// LoginTelarHandlerSPA creates a handler for logging in telar social
func LoginTelarHandlerSPA(c *fiber.Ctx, model *models.LoginModel) error {

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
	}

	if model.Username == "" {
		log.Error("Username is required!")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("usernameIsRequired", "Username is required!"))

	}

	if model.Password == "" {
		log.Error("Password is required!")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("passwordIsRequired", "Password is required!"))
	}

	foundUser, err := userAuthService.FindByUsername(model.Username)
	if err != nil || foundUser == nil {
		if err != nil {
			log.Error(" User not found %s", err.Error())
		}
		log.Error("User not found!")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findUserByUserName", "User not found!"))

	}

	if !foundUser.EmailVerified && !foundUser.PhoneVerified {

		log.Error("User is not verified!")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userNotVerified", "User is not verified!"))
	}

	log.Info(" foundUser.Password: %s  , model.Password: %s", foundUser.Password, model.Password)
	compareErr := utils.CompareHash(foundUser.Password, []byte(model.Password))
	if compareErr != nil {
		log.Error("Password doesn't match %s", compareErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("passwordNotMatch", "Password doesn't match!"))
	}

	profileChannel := readProfileAsync(foundUser.ObjectId)
	langChannel := readLanguageSettingAsync(foundUser.ObjectId,
		&UserInfoInReq{UserId: foundUser.ObjectId, Username: foundUser.Username, SystemRole: foundUser.Role})

	profileResult, langResult := <-profileChannel, <-langChannel
	if profileResult.Error != nil || profileResult.Profile == nil {
		if profileResult.Error != nil {
			log.Error(" User profile  %s", profileResult.Error.Error())
		}
		return c.Status(http.StatusBadRequest).JSON(utils.Error("internal/getUserProfile", "Can not find user profile!"))
	}

	currentUserLang := "en"
	fmt.Println("langResult.settings", langResult.settings)
	langSettigPath := getSettingPath(foundUser.ObjectId, "lang", "current")
	if val, ok := langResult.settings[langSettigPath]; ok && val != "" {
		currentUserLang = val
	} else {
		go func() {
			userInfoReq := &UserInfoInReq{
				UserId:      foundUser.ObjectId,
				Username:    foundUser.Username,
				Avatar:      profileResult.Profile.Avatar,
				DisplayName: profileResult.Profile.FullName,
				SystemRole:  foundUser.Role,
			}
			createDefaultLangSetting(userInfoReq)
		}()
	}

	tokenModel := &TokenModel{
		token:            ProviderAccessToken{},
		oauthProvider:    nil,
		providerName:     "telar",
		profile:          &provider.Profile{Name: foundUser.Username, ID: foundUser.ObjectId.String(), Login: foundUser.Username},
		organizationList: "Red Gold",
		claim: UserClaim{
			DisplayName: profileResult.Profile.FullName,
			SocialName:  profileResult.Profile.SocialName,
			Email:       profileResult.Profile.Email,
			Avatar:      profileResult.Profile.Avatar,
			Banner:      profileResult.Profile.Banner,
			TagLine:     profileResult.Profile.TagLine,
			UserId:      foundUser.ObjectId.String(),
			Role:        foundUser.Role,
			CreatedDate: foundUser.CreatedDate,
		},
	}
	session, err := createToken(tokenModel)
	if err != nil {
		log.Error("Error creating session: %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/createToken", "Internal server error creating token"))
	}

	// Write session on cookie
	writeSessionOnCookie(c, session, &authConfig.AuthConfig)

	// Write user language on cookie
	writeUserLangOnCookie(c, currentUserLang)

	webURL := authConfig.AuthConfig.ExternalRedirectDomain

	redirect := c.Query("r")
	log.Info("SetCookie done, redirect to: %s", redirect)

	// Redirect to original requested resource (if specified in r=)
	if len(redirect) > 0 {
		log.Info(`Found redirect value "r"=%s, instructing client to redirect`, redirect)

		// Note: unable to redirect after setting Cookie, so landing on a redirect page instead.
		webURL = redirect

	}

	return c.JSON(fiber.Map{
		"user":     profileResult.Profile,
		"redirect": webURL,
	})

}

// LoginTelarHandlerSSR creates a handler for logging in telar social
func LoginTelarHandlerSSR(c *fiber.Ctx, model *models.LoginModel) error {

	loginData := &loginPageData{
		title:         "Login - " + *coreConfig.AppConfig.AppName,
		orgName:       *coreConfig.AppConfig.OrgName,
		orgAvatar:     *coreConfig.AppConfig.OrgAvatar,
		appName:       *coreConfig.AppConfig.AppName,
		actionForm:    "",
		resetPassLink: "",
		signupLink:    "",
		message:       "",
	}

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
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

	foundUser, err := userAuthService.FindByUsername(model.Username)
	if err != nil || foundUser == nil {
		if err != nil {
			log.Error(" User not found %s", err.Error())
		}
		loginData.message = "User not found!"
		return loginPageResponse(c, loginData)
	}

	log.Error("[INFO] Found user auth: %v", foundUser)
	if !foundUser.EmailVerified && !foundUser.PhoneVerified {

		loginData.message = "User is not verified!"
		return loginPageResponse(c, loginData)
	}
	log.Info(" foundUser.Password: %s  , model.Password: %s", foundUser.Password, model.Password)
	compareErr := utils.CompareHash(foundUser.Password, []byte(model.Password))
	if compareErr != nil {
		log.Error("Password doesn't match %s", compareErr.Error())
		loginData.message = "Password doesn't match!"
		return loginPageResponse(c, loginData)
	}

	profileChannel := readProfileAsync(foundUser.ObjectId)
	langChannel := readLanguageSettingAsync(foundUser.ObjectId,
		&UserInfoInReq{UserId: foundUser.ObjectId, Username: foundUser.Username, SystemRole: foundUser.Role})

	profileResult, langResult := <-profileChannel, <-langChannel
	if profileResult.Error != nil || profileResult.Profile == nil {
		if profileResult.Error != nil {
			log.Error(" User profile  %s", profileResult.Error.Error())
		}
		loginData.message = "User Profile error!"
		return loginPageResponse(c, loginData)
	}

	currentUserLang := "en"
	fmt.Println("langResult.settings", langResult.settings)
	langSettigPath := getSettingPath(foundUser.ObjectId, "lang", "current")
	if val, ok := langResult.settings[langSettigPath]; ok && val != "" {
		currentUserLang = val
	} else {
		go func() {
			userInfoReq := &UserInfoInReq{
				UserId:      foundUser.ObjectId,
				Username:    foundUser.Username,
				Avatar:      profileResult.Profile.Avatar,
				DisplayName: profileResult.Profile.FullName,
				SystemRole:  foundUser.Role,
			}
			createDefaultLangSetting(userInfoReq)
		}()
	}

	tokenModel := &TokenModel{
		token:            ProviderAccessToken{},
		oauthProvider:    nil,
		providerName:     "telar",
		profile:          &provider.Profile{Name: foundUser.Username, ID: foundUser.ObjectId.String(), Login: foundUser.Username},
		organizationList: "Red Gold",
		claim: UserClaim{
			DisplayName: profileResult.Profile.FullName,
			SocialName:  profileResult.Profile.SocialName,
			Email:       profileResult.Profile.Email,
			Avatar:      profileResult.Profile.Avatar,
			Banner:      profileResult.Profile.Banner,
			TagLine:     profileResult.Profile.TagLine,
			UserId:      foundUser.ObjectId.String(),
			Role:        foundUser.Role,
			CreatedDate: foundUser.CreatedDate,
		},
	}
	session, err := createToken(tokenModel)
	if err != nil {
		log.Error("Error creating session: %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/createToken", "Internal server error creating token"))
	}

	// Write session on cookie
	writeSessionOnCookie(c, session, &authConfig.AuthConfig)

	// Write user language on cookie
	writeUserLangOnCookie(c, currentUserLang)

	webURL := authConfig.AuthConfig.ExternalRedirectDomain

	redirect := c.Query("r")
	log.Info("SetCookie done, redirect to: %s", redirect)

	// Redirect to original requested resource (if specified in r=)
	if len(redirect) > 0 {
		log.Info(`Found redirect value "r"=%s, instructing client to redirect`, redirect)

		// Note: unable to redirect after setting Cookie, so landing on a redirect page instead.

		return c.Render("redirect", fiber.Map{
			"URL": redirect,
		})

	}

	return c.Render("redirect", fiber.Map{
		"URL": webURL,
	})

}

// LoginAdminHandler creates a handler for logging in telar social
func LoginAdminHandler(c *fiber.Ctx) error {

	authConfig := &authConfig.AuthConfig
	// Create the model object
	model := new(models.LoginModel)
	if err := c.BodyParser(model); err != nil {
		errorMessage := fmt.Sprintf("Unmarshal LoginModel Error %s", err.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/modelMarshal", "Can not parse body"))
	}

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
	}

	if model.Username == "" {
		log.Error(" Username is empty")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("usernameRequired", "Username is required!"))

	}

	if model.Password == "" {
		log.Error(" Password is empty")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("passwordRequired", "Password is required!"))

	}

	foundUser, err := userAuthService.FindByUsername(model.Username)
	if err != nil {
		log.Error(" User not found %s", err.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userNotFoundError", "User not found"))

	}

	if foundUser == nil {
		log.Error(" User in null%s", model.Username)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userNotFoundError", "User not found"))
	}

	if !foundUser.EmailVerified && !foundUser.PhoneVerified {
		errorMessage := fmt.Sprintf("User %s is not verified!", foundUser.Username)
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userNotVerifiedError", errorMessage))
	}
	compareErr := utils.CompareHash(foundUser.Password, []byte(model.Password))
	if compareErr != nil {
		log.Error("Password doesn't match %s", compareErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("passwordMatchError", "Password doesn't match "))
	}

	foundUserProfile, errProfile := getUserProfileByID(foundUser.ObjectId)
	if errProfile != nil {
		log.Error(" User profile  %s", errProfile.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findUserProfileError", "Find user profile error"))
	}
	if foundUserProfile == nil {
		log.Error(" User profile is null  %s", foundUser.ObjectId)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findUserProfileError", "Could not find user"))
	}

	tokenModel := &TokenModel{
		token:            ProviderAccessToken{},
		oauthProvider:    nil,
		providerName:     "telar",
		profile:          &provider.Profile{Name: foundUser.Username, ID: foundUser.ObjectId.String(), Login: foundUser.Username},
		organizationList: *coreConfig.AppConfig.OrgName,
		claim: UserClaim{
			DisplayName: foundUserProfile.FullName,
			SocialName:  foundUserProfile.SocialName,
			Email:       foundUserProfile.Email,
			Avatar:      foundUserProfile.Avatar,
			Banner:      foundUserProfile.Banner,
			TagLine:     foundUserProfile.TagLine,
			UserId:      foundUser.ObjectId.String(),
			Role:        foundUser.Role,
			CreatedDate: foundUser.CreatedDate,
		},
	}
	session, err := createToken(tokenModel)
	if err != nil {
		log.Error("Error creating session: %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/createToken", "Internal server error creating token"))
	}

	// Write session on cookie
	writeSessionOnCookie(c, session, authConfig)
	log.Info("Session is created: %s", session)

	return c.JSON(fiber.Map{
		"token": session,
	})

}

// loginPageResponse login page response template
func loginPageResponse(c *fiber.Ctx, data *loginPageData) error {
	return c.Render("login", fiber.Map{
		"Title":         data.title,
		"OrgName":       data.orgName,
		"OrgAvatar":     data.orgAvatar,
		"AppName":       data.appName,
		"ActionForm":    data.actionForm,
		"ResetPassLink": data.resetPassLink,
		"SignupLink":    data.signupLink,
		"GithubLink":    data.githubLink,
		"Message":       data.message,
	})
}

func generateStateOauthCookie(c *fiber.Ctx) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := fiber.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	c.Cookie(&cookie)

	return state
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}
