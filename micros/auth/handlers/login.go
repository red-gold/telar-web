package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	tsconfig "github.com/red-gold/telar-core/config"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	cf "github.com/red-gold/telar-web/micros/auth/config"
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

// LoginHandler creates a handler for logging in
func LoginHandler(server.Request) (handler.Response, error) {
	contents, err := ioutil.ReadFile("./html_template/login.html")
	if err != nil {
		return handler.Response{
			Body: []byte(err.Error()),
		}, err
	}

	return handler.Response{
		Body:       contents,
		StatusCode: http.StatusOK,
	}, nil

}

// LoginGithubHandler creates a handler for logging in github
func LoginGithubHandler(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {

	config := cf.AuthConfig
	log.Println("Login to path", r.URL.Path)

	resource := "/"

	if val := r.URL.Query().Get("r"); len(val) > 0 {
		resource = val
	}

	u := buildGitHubURL(&config, resource, "read:org,read:user,user:email")

	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
	return handler.Response{}, nil

}

// LoginGoogleHandler makes a handler for OAuth 2.0 redirects
func LoginGoogleHandler(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {

	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w)

	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the the state query parameter on your redirect callback.
	*/
	u := googleOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)

	return handler.Response{
		Body: []byte(`You have been issued a cookie. Please navigate to the page you were looking for.`),
	}, nil

}

// LoginPageHandler creates a handler for logging in
func LoginPageHandler(server.Request) (handler.Response, error) {

	appConfig := tsconfig.AppConfig
	authConfig := &cf.AuthConfig
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
	return loginPageResponse(loginData)
}

// LoginTelarHandler creates a handler for logging in telar social
func LoginTelarHandler(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		authConfig := &cf.AuthConfig
		coreConfig := &tsconfig.AppConfig

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

		// Create service
		userAuthService, serviceErr := service.NewUserAuthService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
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

		foundUser, err := userAuthService.FindByUsername(model.Username)
		if err != nil || foundUser.ObjectId == uuid.Nil {
			if err != nil {
				fmt.Printf("\n User not found %s\n", err.Error())
			}
			loginData.message = "User not found!"
			return loginPageResponse(loginData)
		}

		fmt.Printf("[INFO] Found user auth: %v", foundUser)
		if !foundUser.EmailVerified && !foundUser.PhoneVerified {

			loginData.message = "User is not verified!"
			return loginPageResponse(loginData)
		}
		fmt.Printf("\n foundUser.Password: %s  , model.Password: %s", foundUser.Password, model.Password)
		compareErr := utils.CompareHash(foundUser.Password, []byte(model.Password))
		if compareErr != nil {
			fmt.Printf("\nPassword doesn't match %s\n", compareErr.Error())
			loginData.message = "Password doesn't match!"
			return loginPageResponse(loginData)
		}

		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}
		filter := struct {
			ObjectId uuid.UUID `json:"objectId" bson:"objectId"`
		}{
			ObjectId: foundUser.ObjectId,
		}
		extraFOUND, err := userProfileService.FindOneUserProfile(filter)
		fmt.Printf("[INFO] foundUser.ObjectId: %s \n ---->>>>>>, %v", uuid.Must(uuid.FromString(foundUser.ObjectId.String())), extraFOUND)
		foundUserProfile, errProfile := userProfileService.FindByUserId(uuid.Must(uuid.FromString(foundUser.ObjectId.String())))
		if errProfile != nil || foundUserProfile.ObjectId == uuid.Nil {
			if errProfile != nil {
				fmt.Printf("\n User profile  %s\n", errProfile.Error())
			}
			loginData.message = "User Profile error!"
			return loginPageResponse(loginData)
		}
		fmt.Printf("[INFO] Found user profile: %v", foundUserProfile)
		tokenModel := &TokenModel{
			token:            ProviderAccessToken{},
			oauthProvider:    nil,
			providerName:     "telar",
			profile:          &provider.Profile{Name: foundUser.Username, ID: foundUser.ObjectId.String(), Login: foundUser.Username},
			organizationList: "Red Gold",
			claim: UserClaim{
				DisplayName: foundUserProfile.FullName,
				Email:       foundUserProfile.Email,
				Avatar:      foundUserProfile.Avatar,
				UserId:      foundUser.ObjectId.String(),
				Role:        foundUser.Role,
			},
		}
		session, err := createToken(tokenModel)
		if err != nil {
			log.Printf("{error: 'Error creating session: %s'}", err.Error())
			return handler.Response{
				Body:       []byte("{error: 'Internal server error creating JWT'}"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		// Write session on cookie
		writeSessionOnCookie(w, session, authConfig)
		fmt.Printf("\nSession is created: %s \n", session)
		webURL := utils.GetPrettyURLf("/web")

		fmt.Printf("\nwebURL: %s \n", webURL)
		return handler.Response{
			Body:       []byte(`<html><head></head>Redirecting.. <a href="redirect">to original resource</a>. <script>window.location.replace("` + webURL + `");</script></html>`),
			StatusCode: http.StatusOK,
			Header: map[string][]string{
				"Content-Type": {" text/html; charset=utf-8"},
			},
		}, nil

	}
}

// LoginAdminHandler creates a handler for logging in telar social
func LoginAdminHandler(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		authConfig := &cf.AuthConfig
		// Create the model object
		var model models.LoginModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			errorMessage := fmt.Sprintf("Unmarshal LoginModel Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("modelMarshalError", errorMessage)}, nil
		}

		// Create service
		userAuthService, serviceErr := service.NewUserAuthService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		if model.Username == "" {
			fmt.Printf("\n Username is empty\n")
			errorMessage := fmt.Sprintf("Username is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("usernameRequiredError", errorMessage)},
				nil
		}

		if model.Password == "" {
			fmt.Printf("\n Password is empty\n")
			errorMessage := fmt.Sprintf("Password is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("passwordRequiredError", errorMessage)},
				nil
		}

		foundUser, err := userAuthService.FindByUsername(model.Username)
		if err != nil {
			fmt.Printf("\n User not found %s\n", err.Error())
			errorMessage := fmt.Sprintf("User not found %s", err.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userNotFoundError", errorMessage)},
				nil
		}

		if !foundUser.EmailVerified && !foundUser.PhoneVerified {

			errorMessage := fmt.Sprintf("User is not verified!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userNotVerifiedError", errorMessage)},
				nil
		}
		compareErr := utils.CompareHash(foundUser.Password, []byte(model.Password))
		if compareErr != nil {
			fmt.Printf("\nPassword doesn't match %s\n", compareErr.Error())
			errorMessage := fmt.Sprintf("Password doesn't match %s", compareErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("passwordMatchError", errorMessage)},
				nil
		}

		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}
		foundUserProfile, errProfile := userProfileService.FindByUserId(foundUser.ObjectId)
		if errProfile != nil {
			fmt.Printf("\n User profile  %s\n", errProfile.Error())
			errorMessage := fmt.Sprintf("Find user profile %s", errProfile.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("findUserProfileError", errorMessage)},
				nil
		}
		tokenModel := &TokenModel{
			token:            ProviderAccessToken{},
			oauthProvider:    nil,
			providerName:     "telar",
			profile:          &provider.Profile{Name: foundUser.Username, ID: foundUser.ObjectId.String(), Login: foundUser.Username},
			organizationList: "Red Gold",
			claim: UserClaim{
				DisplayName: foundUserProfile.FullName,
				Email:       foundUserProfile.Email,
				Avatar:      foundUserProfile.Avatar,
				UserId:      foundUser.ObjectId.String(),
				Role:        foundUser.Role,
			},
		}
		session, err := createToken(tokenModel)
		if err != nil {
			log.Printf("{error: 'Error creating session: %s'}", err.Error())
			return handler.Response{
				Body:       []byte("{error: 'Internal server error creating JWT'}"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		// Write session on cookie
		writeSessionOnCookie(w, session, authConfig)
		fmt.Printf("\nSession is created: %s \n", session)
		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true, "token": "%s"}`, session)),
			StatusCode: http.StatusOK,
		}, nil
	}
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
		GithubLink    string
		Message       string
	}{
		Title:         data.title,
		OrgName:       data.orgName,
		OrgAvatar:     data.orgAvatar,
		AppName:       data.appName,
		ActionForm:    data.actionForm,
		ResetPassLink: data.resetPassLink,
		SignupLink:    data.signupLink,
		GithubLink:    data.githubLink,
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

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

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
