package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	coreConfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/constants"
	authConfig "github.com/red-gold/telar-web/micros/auth/config"
	"github.com/red-gold/telar-web/micros/auth/database"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
	models "github.com/red-gold/telar-web/micros/auth/models"
	"github.com/red-gold/telar-web/micros/auth/provider"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

// Data for signup verify page template
type signupVerifyPageData struct {
	title      string
	orgName    string
	orgAvatar  string
	appName    string
	actionForm string
	baseRoutes string
	token      string
	message    string
}

// VerifySignupHandle verify signup token
func VerifySignupHandle(c *fiber.Ctx) error {

	model := &models.VerifySignupModel{
		Code:         c.FormValue("code"),
		Token:        c.FormValue("verificaitonSecret"),
		ResponseType: c.FormValue("responseType"),
	}

	if model.ResponseType == "spa" {
		return VerifySignupSPA(c, model)
	}
	return VerifySignupSSR(c, model)

}

func VerifySignupSPA(c *fiber.Ctx, model *models.VerifySignupModel) error {

	// Validate token
	remoteIpAddress := c.IP()

	claims, errToken := utils.ValidateToken([]byte(*coreConfig.AppConfig.PublicKey), model.Token)
	if errToken != nil {
		log.Error("[VerifySignupSPA] Token validation: %s", errToken.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("needValidToken", "Error happened in validating token!"))
	}

	claimMap, _ := claims["claim"].(map[string]interface{})
	userRemoteIp, _ := claimMap["remoteIpAddress"].(string)
	verifyType := claimMap["verifyType"].(string)
	verifyMode, _ := claimMap["mode"].(string)
	verifyId, _ := claimMap["verifyId"].(string)
	userId, _ := claimMap["userId"].(string)
	fullName, _ := claimMap["fullName"].(string)
	email, _ := claimMap["email"].(string)
	phoneNumber, _ := claimMap["phoneNumber"].(string)
	password, _ := claimMap["password"].(string)
	verifyTarget := ""
	fmt.Printf("\nuserId: %s, fullName: %s, email: %s, password: %s, userRemoteIp: %s, verifyType: %v, verifyMode: %v, verifyId: %s\n",
		userId, fullName, email, password, userRemoteIp, verifyType, verifyMode, verifyId)
	emailVerified := false
	phoneVerified := false

	if verifyType == constants.EmailVerifyConst.String() {
		verifyTarget = email
		emailVerified = true
	} else {
		verifyTarget = phoneNumber
		phoneVerified = true
	}
	if remoteIpAddress != userRemoteIp {

		log.Error("[VerifySignupSPA] The request is from different remote ip address!")
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidToken", "Error happened in validating token!"))

	}

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
	}

	userVerificationService, serviceErr := service.NewUserVerificationService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userVerificationService", serviceErr.Error()))
	}

	userUUID, userUuidErr := uuid.FromString(userId)
	if userUuidErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseUserUUIDError", userUuidErr.Error()))
	}

	verifyUUID, verifyUuidErr := uuid.FromString(verifyId)
	if verifyUuidErr != nil {
		errorMessage := fmt.Sprintf("Can not parse verify id! error: %s", verifyUuidErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseVerifyUUIDError", "Can not parse verify id!"))
	}

	verifyStatus, verifyErr := userVerificationService.VerifyUserByCode(userUUID, verifyUUID, remoteIpAddress, model.Code, verifyTarget)
	if verifyErr != nil {
		errorMessage := fmt.Sprintf("Cannot verify user by provided code! error: %s", verifyErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("invalidCode", "Cannot verify user by provided code!"))
	}

	if !verifyStatus {

		errorMessage := "The code is wrong!"
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("wrongCode", "The code is wrong!"))
	}
	createdDate := utils.UTCNowUnix()
	hashPassword, hashErr := utils.Hash(password)
	if hashErr != nil {
		errorMessage := fmt.Sprintf("Cannot hash the password! error: %s", hashErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal", "Error happened during verification!"))
	}

	newUserAuth := &dto.UserAuth{
		ObjectId:      userUUID,
		Username:      email,
		Password:      hashPassword,
		AccessToken:   model.Token,
		EmailVerified: emailVerified,
		Role:          "user",
		PhoneVerified: phoneVerified,
		CreatedDate:   createdDate,
		LastUpdated:   createdDate,
	}
	userAuthErr := userAuthService.SaveUserAuth(newUserAuth)
	if userAuthErr != nil {

		errorMessage := fmt.Sprintf("Cannot save user authentication! error: %s", userAuthErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal", "Error happened during verification!"))
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
		log.Error("Save user profile %s", userProfileErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("canNotSaveUserProfile", "Cannot save user profile!"))
	}

	setupErr := initUserSetup(newUserAuth.ObjectId, newUserAuth.Username, "", newUserProfile.FullName, newUserAuth.Role)
	if setupErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("initUserSetupError", fmt.Sprintf("Cannot initialize user setup! error: %s", setupErr.Error())))
	}

	tokenModel := &TokenModel{
		token:            ProviderAccessToken{},
		oauthProvider:    nil,
		providerName:     *coreConfig.AppConfig.AppName,
		profile:          &provider.Profile{Name: fullName, ID: userId, Login: email},
		organizationList: *coreConfig.AppConfig.OrgName,
		claim: UserClaim{
			DisplayName: fullName,
			Email:       email,
			UserId:      userId,
			Role:        "user",
		},
	}
	session, sessionErr := createToken(tokenModel)
	if sessionErr != nil {
		errorMessage := fmt.Sprintf("Error creating session error: %s",
			sessionErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("initUserSetupError", errorMessage))

	}

	log.Info("\nSession is created: %s \n", session)
	webURL := authConfig.AuthConfig.ExternalRedirectDomain
	return c.Render("redirect", fiber.Map{
		"URL": webURL,
	})
}

func VerifySignupSSR(c *fiber.Ctx, model *models.VerifySignupModel) error {

	prettyURL := utils.GetPrettyURLf(authConfig.AuthConfig.BaseRoute)

	signupVerifyData := &signupVerifyPageData{
		title:      "Login - Telar Social",
		orgName:    *coreConfig.AppConfig.OrgName,
		orgAvatar:  *coreConfig.AppConfig.OrgAvatar,
		appName:    *coreConfig.AppConfig.AppName,
		actionForm: prettyURL + "/signup/verify",
		token:      model.Token,
		message:    "",
	}
	// Validate token
	remoteIpAddress := c.IP()

	claims, errToken := utils.ValidateToken([]byte(*coreConfig.AppConfig.PublicKey), model.Token)
	if errToken != nil {
		errorMessage := fmt.Sprintf("Can not parse token : %s",
			errToken.Error())
		signupVerifyData.message = errorMessage
		return renderCodeVerify(c, signupVerifyData)
	}
	claimMap, _ := claims["claim"].(map[string]interface{})
	userRemoteIp, _ := claimMap["remoteIpAddress"].(string)
	verifyType := claimMap["verifyType"].(string)
	verifyMode, _ := claimMap["mode"].(string)
	verifyId, _ := claimMap["verifyId"].(string)
	userId, _ := claimMap["userId"].(string)
	fullName, _ := claimMap["fullName"].(string)
	email, _ := claimMap["email"].(string)
	phoneNumber, _ := claimMap["phoneNumber"].(string)
	password, _ := claimMap["password"].(string)
	verifyTarget := ""
	fmt.Printf("\nuserId: %s, fullName: %s, email: %s, password: %s, userRemoteIp: %s, verifyType: %v, verifyMode: %v, verifyId: %s\n",
		userId, fullName, email, password, userRemoteIp, verifyType, verifyMode, verifyId)
	emailVerified := false
	phoneVerified := false

	if verifyType == constants.EmailVerifyConst.String() {
		verifyTarget = email
		emailVerified = true
	} else {
		verifyTarget = phoneNumber
		phoneVerified = true
	}
	if remoteIpAddress != userRemoteIp {

		errorMessage := "The request is from different remote ip address!"
		signupVerifyData.message = errorMessage
		return renderCodeVerify(c, signupVerifyData)
	}

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
	}

	userVerificationService, serviceErr := service.NewUserVerificationService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userVerificationService", serviceErr.Error()))
	}

	userUUID, userUuidErr := uuid.FromString(userId)
	if userUuidErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseUserUUIDError", userUuidErr.Error()))
	}

	verifyUUID, verifyUuidErr := uuid.FromString(verifyId)
	if verifyUuidErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseVerifyUUIDError", fmt.Sprintf("Can not parse verify id! error: %s", verifyUuidErr.Error())))
	}

	verifyStatus, verifyErr := userVerificationService.VerifyUserByCode(userUUID, verifyUUID, remoteIpAddress, model.Code, verifyTarget)
	if verifyErr != nil {
		errorMessage := fmt.Sprintf("Cannot verify user by provided code! error: %s", verifyErr.Error())
		signupVerifyData.message = errorMessage
		return renderCodeVerify(c, signupVerifyData)
	}

	if !verifyStatus {

		errorMessage := "The code is wrong!"
		signupVerifyData.message = errorMessage
		return renderCodeVerify(c, signupVerifyData)
	}
	createdDate := utils.UTCNowUnix()
	hashPassword, hashErr := utils.Hash(password)
	if hashErr != nil {
		errorMessage := fmt.Sprintf("Cannot hash the password! error: %s", hashErr.Error())
		signupVerifyData.message = errorMessage
		return renderCodeVerify(c, signupVerifyData)
	}
	newUserAuth := &dto.UserAuth{
		ObjectId:      userUUID,
		Username:      email,
		Password:      hashPassword,
		AccessToken:   model.Token,
		EmailVerified: emailVerified,
		Role:          "user",
		PhoneVerified: phoneVerified,
		CreatedDate:   createdDate,
		LastUpdated:   createdDate,
	}
	userAuthErr := userAuthService.SaveUserAuth(newUserAuth)
	if userAuthErr != nil {

		errorMessage := fmt.Sprintf("Cannot save user authentication! error: %s", userAuthErr.Error())
		signupVerifyData.message = errorMessage
		return renderCodeVerify(c, signupVerifyData)
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
		log.Error("Save user profile %s", userProfileErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("canNotSaveUserProfile", fmt.Sprintf("Cannot save user profile! error: %s", userProfileErr.Error())))
	}

	setupErr := initUserSetup(newUserAuth.ObjectId, newUserAuth.Username, "", newUserProfile.FullName, newUserAuth.Role)
	if setupErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("initUserSetupError", fmt.Sprintf("Cannot initialize user setup! error: %s", setupErr.Error())))
	}

	tokenModel := &TokenModel{
		token:            ProviderAccessToken{},
		oauthProvider:    nil,
		providerName:     *coreConfig.AppConfig.AppName,
		profile:          &provider.Profile{Name: fullName, ID: userId, Login: email},
		organizationList: *coreConfig.AppConfig.OrgName,
		claim: UserClaim{
			DisplayName: fullName,
			Email:       email,
			UserId:      userId,
			Role:        "user",
		},
	}
	session, sessionErr := createToken(tokenModel)
	if sessionErr != nil {
		errorMessage := fmt.Sprintf("Error creating session error: %s",
			sessionErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("initUserSetupError", errorMessage))

	}

	log.Info("\nSession is created: %s \n", session)
	webURL := authConfig.AuthConfig.ExternalRedirectDomain
	return c.Render("redirect", fiber.Map{
		"URL": webURL,
	})
}

// CheckAdminHandler creates a handler to check whether admin user registered
func CheckAdminHandler(c *fiber.Ctx) error {

	//Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
	}

	adminUser, checkErr := userAuthService.CheckAdmin()
	if checkErr != nil {
		errorMessage := fmt.Sprintf("Admin check error: %s",
			checkErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("adminCheckError", errorMessage))
	}

	log.Info("Admin check: %v", adminUser)

	adminExist := (adminUser.ObjectId != uuid.Nil)

	return c.JSON(fiber.Map{
		"admin": adminExist,
	})

}

// renderCodeVerify return signup verify page
func renderCodeVerify(c *fiber.Ctx, data *signupVerifyPageData) error {
	return c.Render("code_verification", fiber.Map{
		"Title":      data.title,
		"OrgName":    data.orgName,
		"OrgAvatar":  data.orgAvatar,
		"AppName":    data.appName,
		"ActionForm": data.actionForm,
		"SignupLink": "",
		"Secret":     data.token,
		"Message":    data.message,
	})
}
