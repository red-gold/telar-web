package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	tsconfig "github.com/red-gold/telar-core/config"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/constants"
	cf "github.com/red-gold/telar-web/micros/auth/config"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

// ResetPasswordPageHandler creates a handler for logging in
func ResetPasswordPageHandler(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		verifyId := req.GetParamByName("verifyId")
		appConfig := tsconfig.AppConfig
		authConfig := cf.AuthConfig

		// Create service
		userAuthService, serviceErr := service.NewUserAuthService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		userVerificationService, serviceErr := service.NewUserVerificationService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		verifyUUID, uuidErr := uuid.FromString(verifyId)
		if uuidErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, uuidErr
		}

		foundVerification, findErr := userVerificationService.FindByVerifyId(verifyUUID)
		if findErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, findErr
		}

		_, userAuthErr := userAuthService.FindByUserId(foundVerification.UserId)
		if userAuthErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, userAuthErr
		}
		prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)

		html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/reset_password.html", struct {
			Title         string
			OrgName       string
			OrgAvatar     string
			AppName       string
			ActionForm    string
			ResetPassLink string
			LoginLink     string
		}{
			Title:         "Login - Telar Social",
			OrgName:       *appConfig.OrgName,
			OrgAvatar:     *appConfig.OrgAvatar,
			AppName:       *appConfig.AppName,
			ActionForm:    fmt.Sprintf("%s/password/reset/%s", prettyURL, verifyId),
			ResetPassLink: "",
			LoginLink:     prettyURL + "/login",
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
}

// ForgetPasswordPageHandler creates a handler for logging in
func ForgetPasswordPageHandler(req server.Request) (handler.Response, error) {
	appConfig := tsconfig.AppConfig
	authConfig := cf.AuthConfig
	loginURL := utils.GetPrettyURLf(authConfig.BaseRoute + "/login")
	html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/forget_password.html", struct {
		Title      string
		OrgName    string
		OrgAvatar  string
		AppName    string
		ActionForm string
		LoginLink  string
	}{
		Title:      "Login - Telar Social",
		OrgName:    *appConfig.OrgName,
		OrgAvatar:  *appConfig.OrgAvatar,
		AppName:    *appConfig.AppName,
		ActionForm: "",
		LoginLink:  loginURL,
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

// ForgetPasswordFormHandler creates a handler for logging in
func ForgetPasswordFormHandler(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	appConfig := tsconfig.AppConfig
	authConfig := cf.AuthConfig
	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		params, parseQueryErr := url.ParseQuery(string(req.Body))

		if parseQueryErr != nil {
			fmt.Printf("Can not parse the data form! error: %s ", parseQueryErr)
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("parseDataFormError", "Can not parse the data form!")},
				nil
		}
		userEmail := params.Get("email")

		// Create service
		userAuthService, serviceErr := service.NewUserAuthService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}
		foundUserAuth, userAuthErr := userAuthService.FindByUsername(userEmail)
		if userAuthErr != nil {
			errorMessage := fmt.Sprintf("User not found: %s",
				userAuthErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userNotFound", errorMessage)},
				nil
		}

		userVerificationService, serviceErr := service.NewUserVerificationService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}
		verifyId, uuidErr := uuid.NewV4()
		if uuidErr != nil {
			fmt.Printf("Error in uuid.NewV4 error: %s", uuidErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("canNotCreateUserId", "Error in creating new user id!")},
				nil
		}

		newUserVerification := &dto.UserVerification{
			ObjectId:        verifyId,
			UserId:          foundUserAuth.ObjectId,
			Code:            "0",
			Target:          foundUserAuth.Username,
			TargetType:      constants.EmailVerifyConst,
			Counter:         1,
			RemoteIpAddress: req.IpAddress,
		}
		saveErr := userVerificationService.SaveUserVerification(newUserVerification)
		if saveErr != nil {
			fmt.Printf("Can not save UserVerification: %s", saveErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("canNotSaveVerification", "Error in preparing verification for reset password!")},
				nil
		}
		// Send email

		email := utils.NewEmail(*appConfig.RefEmail, *appConfig.RefEmailPass, *appConfig.SmtpEmail)
		emailReq := utils.NewEmailRequest([]string{foundUserAuth.Username}, "Reset Password", "")
		prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)

		emailResStatus, emailResErr := email.SendEmail(emailReq, "html_template/email_link_verify_reset_pass.html", struct {
			Name    string
			AppName string
			Link    string
			Email   string
		}{
			Name:    foundUserAuth.Username,
			AppName: *appConfig.AppName,
			Link:    fmt.Sprintf("%s%s/password/reset/%s", *appConfig.Gateway, prettyURL, verifyId),
			Email:   foundUserAuth.Username,
		})

		if emailResErr != nil {
			fmt.Printf("Error happened in sending email error: %s", emailResErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("sendEmailError", "Unable to send email!")},
				nil
		}
		if !emailResStatus {
			fmt.Printf("Email response status is false! ")
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("sendEmailStatusError", "Email response status is false! ")},
				nil
		}
		html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/message.html", struct {
			Title     string
			OrgAvatar string
			Message   string
		}{
			Title:     "Reset Password - Telar Social",
			OrgAvatar: *appConfig.OrgAvatar,
			Message:   fmt.Sprintf("Reset password link has been sent to %s. It may takes up to 30 minutes to receive the email.", userEmail),
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
}

// ResetPasswordFormHandler creates a handler for logging in
func ResetPasswordFormHandler(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	appConfig := tsconfig.AppConfig
	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		verifyId := req.GetParamByName("verifyId")
		params, parseQueryErr := url.ParseQuery(string(req.Body))

		if parseQueryErr != nil {
			fmt.Printf("Can not parse the data form! error: %s ", parseQueryErr)
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("parseDataFormError", "Can not parse the data form!")},
				nil
		}
		newPassword := params.Get("newPassword")
		confirmPassword := params.Get("confirmPassword")

		if newPassword != confirmPassword {
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("passwordNotMatchError", "Confirm password didn't match")}, nil
		}

		// Create service
		userAuthService, serviceErr := service.NewUserAuthService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		userVerificationService, serviceErr := service.NewUserVerificationService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		verifyUUID, uuidErr := uuid.FromString(verifyId)
		if uuidErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, uuidErr
		}

		foundVerification, findErr := userVerificationService.FindByVerifyId(verifyUUID)
		if findErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, findErr
		}

		foundUserAuth, userAuthErr := userAuthService.FindByUserId(foundVerification.UserId)
		if userAuthErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, userAuthErr
		}

		hashPassword, hashErr := utils.Hash(newPassword)
		if hashErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest}, hashErr
		}

		updateErr := userAuthService.UpdatePassword(foundUserAuth.ObjectId, hashPassword)
		if updateErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, updateErr
		}

		html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/message.html", struct {
			Title     string
			OrgAvatar string
			Message   string
		}{
			Title:     "Reset Password - Telar Social",
			OrgAvatar: *appConfig.OrgAvatar,
			Message:   fmt.Sprintf("Your password has been updated. You can login with new password."),
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
}
