package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/alexellis/hmac"
	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	coreConfig "github.com/red-gold/telar-core/config"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/constants"
	ac "github.com/red-gold/telar-web/micros/auth/config"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
	models "github.com/red-gold/telar-web/micros/auth/models"
	"github.com/red-gold/telar-web/micros/auth/provider"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

// SignupHandle sign up user
func SignupHandle(db interface{}) func(server.Request) (handler.Response, error) {
	return func(req server.Request) (handler.Response, error) {
		// Create service
		userAuthService, serviceErr := service.NewUserAuthService(db)
		if serviceErr != nil {

			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		model := models.UserRegisterModel{}
		unmarshalErr := json.Unmarshal(req.Body, &model)
		if unmarshalErr != nil {
			errorMessage := fmt.Sprintf("{error: 'Error while un-marshaling UserRegisterModel: %s'}",
				unmarshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: []byte(errorMessage)},
				unmarshalErr

		}

		if model.Password != model.ConfirmPassword {
			passError := fmt.Errorf(`{"error": "Confirm password didn't match"}`)
			return handler.Response{StatusCode: http.StatusBadRequest, Body: []byte(`{"error": "Confirm password didn't match"}`)}, passError
		}

		hashPassword, hashErr := utils.Hash(model.Password)
		if hashErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest}, hashErr
		}

		userAuth := &dto.UserAuth{
			Username:     model.Username,
			Password:     hashPassword,
			AccessToken:  "",
			TokenExpires: 34567890987654,
		}

		err := userAuthService.SaveUserAuth(userAuth)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		body := userAuth.ObjectId.String()

		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// SignupPageHandler creates a handler for logging in
func SignupPageHandler(server.Request) (handler.Response, error) {

	appConfig := coreConfig.AppConfig
	authConfig := &ac.AuthConfig
	html, parseErr := utils.ParseHtmlBytesTemplate("./html_template/signup.html", struct {
		Title        string
		OrgName      string
		OrgAvatar    string
		AppName      string
		ActionForm   string
		LoginLink    string
		RecaptchaKey string
		VerifyType   string
	}{
		Title:        "Signup - Telar Social",
		OrgName:      *appConfig.OrgName,
		OrgAvatar:    *appConfig.OrgAvatar,
		AppName:      *appConfig.AppName,
		ActionForm:   "",
		LoginLink:    "",
		RecaptchaKey: *appConfig.RecaptchaSiteKey,
		VerifyType:   authConfig.VerifyType,
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

// SignupTokenHandle create signup token
func SignupTokenHandle(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		config := coreConfig.AppConfig
		authConfig := &ac.AuthConfig
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

		model := &models.SignupTokenModel{
			User: models.UserSignupTokenModel{
				Fullname: query.Get("fullName"),
				Email:    query.Get("email"),
				Password: query.Get("newPassword"),
			},
			VerifyType: query.Get("verifyType"),
			Recaptcha:  query.Get("g-recaptcha-response"),
		}

		// Verify Captha

		recaptcha := utils.NewRecaptha(*config.RecaptchaKey)
		remoteIpAddress := utils.GetIPAdress(r)
		recaptchaStatus, recaptchaErr := recaptcha.VerifyCaptch(model.Recaptcha, remoteIpAddress)
		if recaptchaErr != nil {
			fmt.Printf("Can not verify recaptcha %s error: %s", *config.RecaptchaKey, recaptchaErr)
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: []byte("{error: 'Error happened in verifying captcha!'}")},
				recaptchaErr
		}
		if !recaptchaStatus {
			fmt.Printf("Error happened in validating recaptcha!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: []byte("{error: 'Recaptcha is not valid!'}")},
				nil
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

		// Check user exist
		userAuth, findError := userAuthService.FindByUsername(model.User.Email)
		if findError != nil {
			errorMessage := fmt.Sprintf("Error while finding user by user name : %s",
				findError.Error())
			fmt.Println(errorMessage)

		}

		if userAuth.ObjectId != uuid.Nil {
			errorMessage := fmt.Sprintf(`{"error": "Error user already exist %s"}`, model.User.Email)
			return handler.Response{StatusCode: http.StatusBadGateway, Body: []byte(errorMessage)},
				findError

		}

		// Create signup token
		newUserId, err := uuid.NewV4()
		if err != nil {
			fmt.Printf("Error in uuid.NewV4 error: %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("canNotCreateUserId", "Error in creating new user id!")},
				nil
		}
		token := ""
		var tokenErr error
		if model.VerifyType == constants.EmailVerifyConst.String() {
			token, tokenErr = userVerificationService.CreateEmailVerficationToken(service.EmailVerificationToken{
				UserId:          newUserId,
				HtmlTmplPath:    "html_template/email_code_verify.html",
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
			fmt.Printf("Error on creating token: %s", tokenErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: []byte(`{"error" : "Error happened in creating token! ` + tokenErr.Error() + `"}`)},
				nil
		}

		// Parse code verification page
		appConfig := coreConfig.AppConfig
		prettyURL := utils.GetPrettyURLf(authConfig.BaseRoute)

		signupVerifyData := &signupVerifyPageData{
			title:      "Login - Telar Social",
			orgName:    *appConfig.OrgName,
			orgAvatar:  *appConfig.OrgAvatar,
			appName:    *appConfig.AppName,
			actionForm: prettyURL + "/signup/verify",
			token:      token,
			message:    "",
		}

		return codeVerifyResponsePage(signupVerifyData)

	}
}

// AdminSignupHandle verify signup token
func AdminSignupHandle(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		authConfig := &ac.AuthConfig
		fullName := "admin"

		email := authConfig.AdminUsername
		password := authConfig.AdminPassword

		// Create service
		userAuthService, serviceErr := service.NewUserAuthService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		userUUID, userUuidErr := uuid.NewV4()
		if userUuidErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("newUserUUIDError", fmt.Sprintf("Can not create user id! error: %s", userUuidErr.Error()))},
				userUuidErr
		}

		createdDate := utils.UTCNowUnix()
		hashPassword, hashErr := utils.Hash(password)
		if hashErr != nil {
			errorMessage := fmt.Sprintf("Cannot hash the password! error: %s", hashErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("hashPasswordError", errorMessage)},
				nil
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
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("saveUserAuthError", errorMessage)},
				nil
		}

		newUserProfile := &dto.UserProfile{
			ObjectId:    userUUID,
			FullName:    fullName,
			CreatedDate: createdDate,
			LastUpdated: createdDate,
			Email:       email,
			Avatar:      "https://util.telar.dev/api/avatars/" + userUUID.String(),
			Banner:      fmt.Sprintf("https://picsum.photos/id/%d/900/300/?blur", generateRandomNumber(1, 1000)),
			Permission:  constants.Public,
		}
		userProfileErr := userProfileService.SaveUserProfile(newUserProfile)
		if userProfileErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError,
					Body: utils.MarshalError("canNotSaveUserProfile",
						fmt.Sprintf("Cannot save user profile! error: %s", userProfileErr.Error()))},
				userProfileErr

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
			profile:          &provider.Profile{Name: fullName, ID: userUUID.String(), Login: email},
			organizationList: "Red Gold",
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
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("readPrivateError", errorMessage)},
				nil
		}

		fmt.Printf("\nSession is created: %s \n", session)

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true, "token": "%s"}`, session)),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// functionCall send request to another function/microservice using HMAC validation
func functionCall(bytesReq []byte, url string) ([]byte, error) {
	prettyURL := utils.GetPrettyURLf(url)
	bodyReader := bytes.NewBuffer(bytesReq)

	httpReq, httpErr := http.NewRequest(http.MethodPost, *coreConfig.AppConfig.Gateway+prettyURL, bodyReader)
	if httpErr != nil {
		return nil, httpErr
	}

	payloadSecret, psErr := utils.ReadSecret("payload-secret")

	if psErr != nil {
		return nil, fmt.Errorf("couldn't get payload-secret: %s", psErr.Error())
	}

	digest := hmac.Sign(bytesReq, []byte(payloadSecret))
	httpReq.Header.Set("Content-type", "application/json")
	fmt.Printf("\ndigest: %s, header: %v \n", "sha1="+hex.EncodeToString(digest), server.X_Cloud_Signature)
	httpReq.Header.Add(server.X_Cloud_Signature, "sha1="+hex.EncodeToString(digest))

	c := http.Client{}
	res, reqErr := c.Do(httpReq)
	fmt.Printf("\nRes: %v\n", res)
	if reqErr != nil {
		return nil, fmt.Errorf("Error while sending admin check request!: %s", reqErr.Error())
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	resData, readErr := ioutil.ReadAll(res.Body)
	if resData == nil || readErr != nil {
		return nil, fmt.Errorf("failed to read response from admin check request.")
	}

	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to call %s api, invalid status: %s", prettyURL, res.Status)
	}

	return resData, nil
}
