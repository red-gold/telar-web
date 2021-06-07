package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	coreConfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/log"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/constants"
	ac "github.com/red-gold/telar-web/micros/auth/config"
	"github.com/red-gold/telar-web/micros/auth/database"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
	models "github.com/red-gold/telar-web/micros/auth/models"
	"github.com/red-gold/telar-web/micros/auth/provider"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

// SignupPageHandler creates a handler for logging in
func SignupPageHandler(c *fiber.Ctx) error {

	appConfig := coreConfig.AppConfig
	authConfig := &ac.AuthConfig
	prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)

	return c.Render("signup", fiber.Map{
		"Title":        "Signup - Telar Social",
		"OrgName":      *appConfig.OrgName,
		"OrgAvatar":    *appConfig.OrgAvatar,
		"AppName":      *appConfig.AppName,
		"ActionForm":   "",
		"LoginLink":    prettyURL + "/login",
		"RecaptchaKey": *appConfig.RecaptchaSiteKey,
		"VerifyType":   authConfig.VerifyType,
	})
}

// SignupTokenHandle create signup token
func SignupTokenHandle(c *fiber.Ctx) error {
	config := coreConfig.AppConfig
	authConfig := &ac.AuthConfig

	model := &models.SignupTokenModel{
		User: models.UserSignupTokenModel{
			Fullname: c.FormValue("fullName"),
			Email:    c.FormValue("email"),
			Password: c.FormValue("newPassword"),
		},
		VerifyType:   c.FormValue("verifyType"),
		Recaptcha:    c.FormValue("g-recaptcha-response"),
		ResponseType: c.FormValue("responseType"),
	}

	// Verify Captha
	recaptcha := utils.NewRecaptha(*config.RecaptchaKey)
	remoteIpAddress := c.IP()
	recaptchaStatus, recaptchaErr := recaptcha.VerifyCaptch(model.Recaptcha, remoteIpAddress)
	if recaptchaErr != nil {
		log.Error("Can not verify recaptcha %s error: %s", *config.RecaptchaKey, recaptchaErr)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/recaptcha", "Error happened in verifying captcha!"))
	}

	if !recaptchaStatus {
		log.Error("Error happened in validating recaptcha!")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("internal/recaptchaNotValid", "Recaptcha is not valid!"))

	}

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/recaptcha", serviceErr.Error()))
	}

	userVerificationService, serviceErr := service.NewUserVerificationService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/recaptcha", serviceErr.Error()))
	}

	// Check user exist
	userAuth, findError := userAuthService.FindByUsername(model.User.Email)
	if findError != nil {
		errorMessage := fmt.Sprintf("Error while finding user by user name : %s",
			findError.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findUserAuth", errorMessage))

	}

	if userAuth != nil {
		err := utils.Error("userAlreadyExist", "User already exist - "+model.User.Email)
		return c.Status(http.StatusBadRequest).JSON(err)
	}

	// Create signup token
	newUserId := uuid.Must(uuid.NewV4())

	token := ""
	var tokenErr error
	if model.VerifyType == constants.EmailVerifyConst.String() {
		token, tokenErr = userVerificationService.CreateEmailVerficationToken(service.EmailVerificationToken{
			UserId:          newUserId,
			HtmlTmplPath:    "views/email_code_verify.html",
			Username:        model.User.Email,
			EmailTo:         model.User.Email,
			EmailSubject:    "Your verification code",
			RemoteIpAddress: remoteIpAddress,
			FullName:        model.User.Fullname,
			UserPassword:    model.User.Password,
		}, &config)
	} else if model.VerifyType == constants.PhoneVerifyConst.String() {
		token, tokenErr = userVerificationService.CreatePhoneVerficationToken(service.PhoneVerificationToken{
			UserId:          newUserId,
			Username:        model.User.Email,
			UserEmail:       model.User.Email,
			RemoteIpAddress: remoteIpAddress,
			FullName:        model.User.Fullname,
			UserPassword:    model.User.Password,
		}, &config)
	}
	if tokenErr != nil {
		log.Error("Error on creating token: %s", tokenErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/findUserAuth", "Error happened in creating token!"))
	}

	// Parse code verification page
	appConfig := coreConfig.AppConfig
	prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)

	if model.ResponseType == "spa" {
		return c.JSON(fiber.Map{
			"token": token,
		})
	}

	signupVerifyData := &signupVerifyPageData{
		title:      "Login - Telar Social",
		orgName:    *appConfig.OrgName,
		orgAvatar:  *appConfig.OrgAvatar,
		appName:    *appConfig.AppName,
		actionForm: prettyURL + "/signup/verify",
		token:      token,
		message:    "",
	}

	return renderCodeVerify(c, signupVerifyData)

}

// AdminSignupHandle verify signup token
func AdminSignupHandle(c *fiber.Ctx) error {
	authConfig := &ac.AuthConfig
	fullName := "admin"

	email := authConfig.AdminUsername
	password := authConfig.AdminPassword

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		log.Error(serviceErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", "Error happened when initialize user auth service!"))

	}

	userUUID := uuid.Must(uuid.NewV4())

	createdDate := utils.UTCNowUnix()
	hashPassword, hashErr := utils.Hash(password)
	if hashErr != nil {
		errorMessage := fmt.Sprintf("Cannot hash the password! error: %s", hashErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("hashPassword", "Cannot hash the password!"))
	}
	newUserAuth := &dto.UserAuth{
		ObjectId:      userUUID,
		Username:      email,
		Password:      hashPassword,
		AccessToken:   "",
		Role:          "admin",
		EmailVerified: true,
		PhoneVerified: true,
		CreatedDate:   createdDate,
		LastUpdated:   createdDate,
	}
	userAuthErr := userAuthService.SaveUserAuth(newUserAuth)
	if userAuthErr != nil {
		errorMessage := fmt.Sprintf("Cannot save user authentication! error: %s", userAuthErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("saveUserAuthError", "Cannot save user authentication!"))
	}

	newUserProfile := &models.UserProfileModel{
		ObjectId:    userUUID,
		FullName:    fullName,
		CreatedDate: createdDate,
		LastUpdated: createdDate,
		Email:       email,
		Avatar:      "https://util.telar.dev/api/avatars/" + userUUID.String(),
		Banner:      fmt.Sprintf("https://picsum.photos/id/%d/900/300/?blur", generateRandomNumber(1, 1000)),
		Permission:  constants.Public,
	}
	userProfileErr := saveUserProfile(newUserProfile)
	if userProfileErr != nil {
		log.Error(fmt.Sprintf("Cannot save user profile! error: %s", userProfileErr.Error()))
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("canNotSaveUserProfile", "Cannot save user profile!"))
	}

	setupErr := initUserSetup(newUserAuth.ObjectId, newUserAuth.Username, "", newUserProfile.FullName, newUserAuth.Role)
	if setupErr != nil {
		log.Error(fmt.Sprintf("Cannot initialize user setup! error: %s", setupErr.Error()))
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("canNotSaveUserProfile", "Cannot initialize user setup!"))
	}

	tokenModel := &TokenModel{
		token:            ProviderAccessToken{},
		oauthProvider:    nil,
		providerName:     "telar",
		profile:          &provider.Profile{Name: fullName, ID: userUUID.String(), Login: email},
		organizationList: "Telar",
		claim: UserClaim{
			DisplayName: fullName,
			Email:       email,
			UserId:      userUUID.String(),
			Role:        "admin",
		},
	}
	session, sessionErr := createToken(tokenModel)
	if sessionErr != nil {
		errorMessage := fmt.Sprintf("Error creating session error: %s",
			sessionErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("createToken", "Error while creating session!"))
	}

	log.Info("\nSession is created: %s \n", session)

	return c.JSON(fiber.Map{
		"token": session,
	})

}
