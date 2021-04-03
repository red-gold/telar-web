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
func VerifySignupHandle(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		appConfig := tsconfig.AppConfig
		authConfig := cf.AuthConfig

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

		model := &models.VerifySignupModel{
			Code:  query.Get("code"),
			Token: query.Get("verificaitonSecret"),
		}
		prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)

		signupVerifyData := &signupVerifyPageData{
			title:      "Login - Telar Social",
			orgName:    *appConfig.OrgName,
			orgAvatar:  *appConfig.OrgAvatar,
			appName:    *appConfig.AppName,
			actionForm: prettyURL + "/signup/verify",
			token:      model.Token,
			message:    "",
		}
		// Validate token
		remoteIpAddress := utils.GetIPAdress(r)

		claims, errToken := utils.ValidateToken([]byte(*appConfig.PublicKey), model.Token)
		if errToken != nil {
			errorMessage := fmt.Sprintf("Can not parse token : %s",
				errToken.Error())
			signupVerifyData.message = errorMessage
			return codeVerifyResponsePage(signupVerifyData)
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
			return codeVerifyResponsePage(signupVerifyData)
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

		userUUID, userUuidErr := uuid.FromString(userId)
		if userUuidErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("parseUserUUIDError", fmt.Sprintf("Can not parse user id! error: %s", userUuidErr.Error()))},
				userUuidErr
		}
		verifyUUID, verifyUuidErr := uuid.FromString(verifyId)
		if verifyUuidErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("parseVerifyUUIDError", fmt.Sprintf("Can not parse verify id! error: %s", verifyUuidErr.Error()))},
				nil
		}
		verifyStatus, verifyErr := userVerificationService.VerifyUserByCode(userUUID, verifyUUID, remoteIpAddress, model.Code, verifyTarget)
		if verifyErr != nil {
			errorMessage := fmt.Sprintf("Cannot verify user by provided code! error: %s", verifyErr.Error())
			signupVerifyData.message = errorMessage
			return codeVerifyResponsePage(signupVerifyData)
		}

		if !verifyStatus {

			errorMessage := "The code is wrong!"
			signupVerifyData.message = errorMessage
			return codeVerifyResponsePage(signupVerifyData)
		}
		createdDate := utils.UTCNowUnix()
		hashPassword, hashErr := utils.Hash(password)
		if hashErr != nil {
			errorMessage := fmt.Sprintf("Cannot hash the password! error: %s", hashErr.Error())
			signupVerifyData.message = errorMessage
			return codeVerifyResponsePage(signupVerifyData)
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
			return codeVerifyResponsePage(signupVerifyData)
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
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("canNotSaveUserProfile",
						fmt.Sprintf("Cannot save user profile! error: %s", userProfileErr.Error()))},
				nil

		}
		setupErr := initUserSetup(newUserAuth.ObjectId, newUserAuth.Username, "", newUserProfile.FullName, newUserAuth.Role)
		if setupErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("initUserSetupError",
						fmt.Sprintf("Cannot initialize user setup! error: %s", setupErr.Error()))},
				userProfileErr
		}

		tokenModel := &TokenModel{
			token:            ProviderAccessToken{},
			oauthProvider:    nil,
			providerName:     "telar",
			profile:          &provider.Profile{Name: fullName, ID: userId, Login: email},
			organizationList: "Red Gold",
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
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("readPrivateError", errorMessage)},
				nil
		}

		fmt.Printf("\nSession is created: %s \n", session)
		webURL := authConfig.ExternalRedirectDomain
		return handler.Response{
			Body:       []byte(`<html><head></head>Redirecting.. <a href="redirect">to original resource</a>. <script>window.location.replace("` + webURL + `");</script></html>`),
			StatusCode: http.StatusOK,
			Header: map[string][]string{
				"Content-Type": {" text/html; charset=utf-8"},
			},
		}, nil
	}
}

// CheckAdminHandler creates a handler to check whether admin user registered
func CheckAdminHandler(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		// Create service
		userAuthService, serviceErr := service.NewUserAuthService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		adminUser, checkErr := userAuthService.CheckAdmin()
		if checkErr != nil {
			errorMessage := fmt.Sprintf("Admin check error: %s",
				checkErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("adminCheckError", errorMessage)},
				nil
		}
		fmt.Printf("Admin check: %v", adminUser)
		adminExist := (adminUser.ObjectId != uuid.Nil)
		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true, "admin": %t}`, adminExist)),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// codeVerifyResponsePage return signup verify page
func codeVerifyResponsePage(data *signupVerifyPageData) (handler.Response, error) {
	html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/code_verification.html", struct {
		Title      string
		OrgName    string
		OrgAvatar  string
		AppName    string
		ActionForm string
		SignupLink string
		Secret     string
		Message    string
	}{
		Title:      data.title,
		OrgName:    data.orgName,
		OrgAvatar:  data.orgAvatar,
		AppName:    data.appName,
		ActionForm: data.actionForm,
		SignupLink: "",
		Secret:     data.token,
		Message:    data.message,
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
