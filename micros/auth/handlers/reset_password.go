package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	tsconfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/constants"
	cf "github.com/red-gold/telar-web/micros/auth/config"
	"github.com/red-gold/telar-web/micros/auth/database"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
	"github.com/red-gold/telar-web/micros/auth/models"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

// ResetPasswordPageHandler creates a handler for logging in
func ResetPasswordPageHandler(c *fiber.Ctx) error {
	verifyId := c.Params("verifyId")
	appConfig := tsconfig.AppConfig
	authConfig := cf.AuthConfig

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
	}

	userVerificationService, serviceErr := service.NewUserVerificationService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userVerificationService", serviceErr.Error()))
	}

	verifyUUID, uuidErr := uuid.FromString(verifyId)
	if uuidErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseVerifyUUIDError", uuidErr.Error()))

	}

	foundVerification, findErr := userVerificationService.FindByVerifyId(verifyUUID)
	if findErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findVerification", findErr.Error()))
	}

	foundUserAuth, userAuthErr := userAuthService.FindByUserId(foundVerification.UserId)
	if userAuthErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findUserAuth", userAuthErr.Error()))
	}

	if foundUserAuth == nil {
		return c.Status(http.StatusNotFound).JSON(utils.Error("notFoundUser", "Could not find user with veridy ID "+verifyId))
	}

	prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)

	return c.Render("reset_password", fiber.Map{
		"Title":         "Login - " + *appConfig.AppName,
		"OrgName":       *appConfig.OrgName,
		"OrgAvatar":     *appConfig.OrgAvatar,
		"AppName":       *appConfig.AppName,
		"ActionForm":    fmt.Sprintf("%s/password/reset/%s", prettyURL, verifyId),
		"ResetPassLink": "",
		"LoginLink":     prettyURL + "/login",
	})

}

// ForgetPasswordPageHandler creates a handler for logging in
func ForgetPasswordPageHandler(c *fiber.Ctx) error {
	appConfig := tsconfig.AppConfig
	authConfig := cf.AuthConfig
	loginURL := utils.GetPrettyURLf(authConfig.BaseRoute + "/login")

	return c.Render("reset_password", fiber.Map{
		"Title":      "Login - " + *appConfig.AppName,
		"OrgName":    *appConfig.OrgName,
		"OrgAvatar":  *appConfig.OrgAvatar,
		"AppName":    *appConfig.AppName,
		"ActionForm": "",
		"LoginLink":  loginURL,
	})

}

// ForgetPasswordFormHandler
func ForgetPasswordFormHandler(c *fiber.Ctx) error {
	appConfig := tsconfig.AppConfig
	authConfig := cf.AuthConfig

	userEmail := c.FormValue("email")
	responseType := c.FormValue("responseType")

	if userEmail == "" {
		errorMessage := fmt.Sprintf("Email is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("emailIsRequired", errorMessage))

	}

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
	}

	foundUserAuth, userAuthErr := userAuthService.FindByUsername(userEmail)
	if userAuthErr != nil {
		errorMessage := fmt.Sprintf("User not found: %s",
			userAuthErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userNotFound", "User not found"))

	}
	if foundUserAuth == nil {
		errorMessage := fmt.Sprintf("User auth not found %s", userEmail)
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userAuthNotFound", "User auth not found"))
	}

	userVerificationService, serviceErr := service.NewUserVerificationService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userVerificationService", serviceErr.Error()))

	}

	verifyId := uuid.Must(uuid.NewV4())

	newUserVerification := &dto.UserVerification{
		ObjectId:        verifyId,
		UserId:          foundUserAuth.ObjectId,
		Code:            "0",
		Target:          foundUserAuth.Username,
		TargetType:      constants.EmailVerifyConst,
		Counter:         1,
		RemoteIpAddress: c.IP(),
	}
	saveErr := userVerificationService.SaveUserVerification(newUserVerification)
	if saveErr != nil {
		log.Error("Can not save UserVerification: %s", saveErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("canNotSaveVerification", "Error in preparing verification for reset password!"))

	}

	// Send email
	email := utils.NewEmail(*appConfig.RefEmail, *appConfig.RefEmailPass, *appConfig.SmtpEmail)
	emailReq := utils.NewEmailRequest([]string{foundUserAuth.Username}, "Reset Password", "")
	prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)

	// Generate reset password token
	token, err := generateResetPasswordToken(verifyId.String())
	if err != nil {
		log.Error("Generate reset password token: %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("generateToken", "Error in generating token!"))

	}

	emailResStatus, emailResErr := email.SendEmail(emailReq, "views/email_link_verify_reset_pass.html", struct {
		Name      string
		AppName   string
		AppURL    string
		Link      string
		Email     string
		OrgName   string
		OrgAvatar string
	}{
		Name:      foundUserAuth.Username,
		AppName:   *appConfig.AppName,
		AppURL:    authConfig.WebURL,
		Link:      fmt.Sprintf("%s%s/password/reset/%s", authConfig.AuthWebURI, prettyURL, token),
		Email:     foundUserAuth.Username,
		OrgName:   *appConfig.OrgName,
		OrgAvatar: *appConfig.OrgAvatar,
	})

	if emailResErr != nil {
		log.Error("Error happened in sending email error: %s", emailResErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("sendEmailError", "Unable to send email!"))

	}

	if !emailResStatus {
		log.Error("Email response status is false! ")
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("sendEmailStatusError", "Email response status is false!"))
	}

	if responseType == SPAResponseType {
		return c.SendStatus(http.StatusOK)
	}

	return c.Render("message", fiber.Map{
		"Title":     "Reset Password - " + *appConfig.AppName,
		"OrgAvatar": *appConfig.OrgAvatar,
		"Message":   fmt.Sprintf("Reset password link has been sent to %s. It may takes up to 30 minutes to receive the email.", userEmail),
	})

}

// ResetPasswordFormHandler creates a handler for logging in
func ResetPasswordFormHandler(c *fiber.Ctx) error {
	appConfig := tsconfig.AppConfig

	verifyIdToken := c.Params("verifyId")
	responseType := c.FormValue("responseType")

	claims, err := decodeResetPasswordToken(verifyIdToken)
	if err != nil {
		log.Error("Can not veify reset password token: %s - %s", verifyIdToken, err.Error())
		return c.Status(http.StatusUnauthorized).JSON(utils.Error("invalidToken", "Can not veify reset password token!"))

	}

	newPassword := c.FormValue("newPassword")
	confirmPassword := c.FormValue("confirmPassword")

	if newPassword != confirmPassword {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("passwordNotMatchError", "Confirm password didn't match"))
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

	verifyUUID, uuidErr := uuid.FromString(claims.VerifyId)
	if uuidErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("parseVerifyUUIDError", uuidErr.Error()))
	}

	foundVerification, findErr := userVerificationService.FindByVerifyId(verifyUUID)
	if findErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findVerification", findErr.Error()))
	}

	foundUserAuth, userAuthErr := userAuthService.FindByUserId(foundVerification.UserId)
	if userAuthErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findUserAuth", userAuthErr.Error()))
	}
	if foundUserAuth == nil {
		errorMessage := fmt.Sprintf("User auth not found %s", foundVerification.UserId)
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userAuthNotFound", "User auth not found"))
	}

	hashPassword, hashErr := utils.Hash(newPassword)
	if hashErr != nil {
		log.Error("Hash password %s", hashErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/hash", "Hash error!"))

	}

	updateErr := userAuthService.UpdatePassword(foundUserAuth.ObjectId, hashPassword)
	if updateErr != nil {
		log.Error("Update user password %s", updateErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/updateUserPassword", "Can not update password!"))
	}

	if responseType == SPAResponseType {
		return c.SendStatus(http.StatusOK)
	}

	return c.Render("message", fiber.Map{
		"Title":     "Reset Password - " + *appConfig.AppName,
		"OrgAvatar": *appConfig.OrgAvatar,
		"Message":   fmt.Sprintf("Your password has been updated. You can login with new password."),
	})

}

// ChangePasswordHandler creates a handler for logging in
func ChangePasswordHandler(c *fiber.Ctx) error {

	model := new(models.ChangePasswordModel)
	unmarshalErr := c.BodyParser(model)
	if unmarshalErr != nil {
		errorMessage := fmt.Sprintf("Error while un-marshaling ChangePasswordModel: %s",
			unmarshalErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/parseModel", "Error while parsing body"))

	}

	if model.NewPassword == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("newPasswordIsRequired", "New password is required!"))
	}

	if model.CurrentPassword == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("currentPasswordIsRequired", "Current password is required!"))
	}

	if model.NewPassword != model.ConfirmPassword {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("passwordNotMatchError", "Confirm password didn't match"))
	}

	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(database.Db)
	if serviceErr != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/userAuthService", serviceErr.Error()))
	}
	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/getCurrentUser", "Can not get current user!"))
	}

	foundUserAuth, userAuthErr := userAuthService.FindByUserId(currentUser.UserID)
	if userAuthErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.Error("findUserAuth", userAuthErr.Error()))
	}

	compareErr := utils.CompareHash(foundUserAuth.Password, []byte(model.CurrentPassword))
	if compareErr != nil {
		log.Error("Current password doesn't match %s", compareErr.Error())
		return c.Status(http.StatusBadRequest).JSON(utils.Error("currentPasswordNotMatch", "Current password doesn't match!"))
	}

	if foundUserAuth == nil {
		errorMessage := fmt.Sprintf("User auth not found %s", currentUser.UserID)
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("userAuthNotFound", "User auth not found"))
	}

	hashPassword, hashErr := utils.Hash(model.NewPassword)
	if hashErr != nil {
		log.Error("Hash password %s", hashErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/hash", "Hash error!"))

	}

	updateErr := userAuthService.UpdatePassword(foundUserAuth.ObjectId, hashPassword)
	if updateErr != nil {
		log.Error("Update user password %s", updateErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/updateUserPassword", "Can not update password!"))
	}

	return c.SendStatus(http.StatusOK)

}
