package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	coreConfig "github.com/red-gold/telar-core/config"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/constants"
	cf "github.com/red-gold/telar-web/micros/auth/config"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
	"github.com/red-gold/telar-web/micros/auth/provider"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

const profileFetchTimeout = time.Second * 5

// checkSignup check for user signup in the case user does not exist in user auth
func checkSignup(accessToken string, model *TokenModel, db interface{}) error {

	if model.profile.Name == "" {
		fmt.Println("[ERROR]: OAuth provide - name can not be empty")
		return fmt.Errorf("OAuth provide - name can not be empty")
	}
	if model.profile.Email == "" {
		fmt.Println("[ERROR]: OAuth provide - email can not be empty")
		return fmt.Errorf("OAuth provide - email can not be empty")
	}
	// Create service
	userAuthService, serviceErr := service.NewUserAuthService(db)
	if serviceErr != nil {
		return serviceErr
	}

	userProfileService, serviceErr := service.NewUserProfileService(db)
	if serviceErr != nil {
		return serviceErr
	}

	fmt.Printf("[INFO]: Oauth check signup for user %s", model.profile.Email)

	// Check user exist
	userAuth, findError := userAuthService.FindByUsername(model.profile.Email)
	if findError != nil {
		errorMessage := fmt.Sprintf("[ERROR]: Error while finding user by user name : %s",
			findError.Error())
		fmt.Println(errorMessage)

	}
	fmt.Printf("[INFO]: Oauth check signup - user auth object %v", userAuth)

	if userAuth.ObjectId == uuid.Nil {
		// Create signup token
		newUserId, uuidErr := uuid.NewV4()
		if uuidErr != nil {
			fmt.Printf("[Error]: uuid.NewV4 error: %s", uuidErr.Error())
			return uuidErr
		}
		createdDate := utils.UTCNowUnix()

		newUserAuth := &dto.UserAuth{
			ObjectId:      newUserId,
			Username:      model.profile.Email,
			Password:      []byte(""),
			AccessToken:   accessToken,
			EmailVerified: true,
			Role:          "user",
			PhoneVerified: false,
			CreatedDate:   createdDate,
			LastUpdated:   createdDate,
		}

		userAuthErr := userAuthService.SaveUserAuth(newUserAuth)
		if userAuthErr != nil {
			return userAuthErr
		}
		model.profile.ID = newUserId.String()
		newUserProfile := &dto.UserProfile{
			ObjectId:    newUserId,
			FullName:    model.profile.Name,
			CreatedDate: createdDate,
			LastUpdated: createdDate,
			Email:       model.profile.Email,
			Avatar:      model.profile.Avatar,
			Banner:      fmt.Sprintf("https://picsum.photos/id/%d/900/300/?blur", generateRandomNumber(1, 1000)),
			Permission:  constants.Public,
		}
		userProfileErr := userProfileService.SaveUserProfile(newUserProfile)
		if userProfileErr != nil {

			return fmt.Errorf("Cannot save user profile! error: %s", userProfileErr.Error())
		}
		setupErr := initUserSetup(newUserAuth.ObjectId, newUserAuth.Username, "", newUserProfile.FullName, newUserAuth.Role)
		if setupErr != nil {
			return fmt.Errorf("Cannot initialize user setup! error: %s", setupErr.Error())
		}
		model.profile.ID = newUserAuth.ObjectId.String()
		model.claim = UserClaim{
			DisplayName: newUserProfile.FullName,
			Email:       newUserProfile.Email,
			UserId:      newUserAuth.ObjectId.String(),
			Role:        newUserAuth.Role,
			Avatar:      newUserProfile.Avatar,
		}
	} else {

		fmt.Printf("\n[INFO]: Check signup user exist, preparing user profile.\n")
		userProfileService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return serviceErr
		}
		foundUserProfile, errProfile := userProfileService.FindByUserId(userAuth.ObjectId)
		if errProfile != nil {
			fmt.Printf("\n User profile  %s\n", errProfile.Error())
			return errProfile
		}

		model.profile.ID = userAuth.ObjectId.String()
		model.profile.Email = foundUserProfile.Email
		model.profile.Name = foundUserProfile.FullName
		model.profile.Avatar = foundUserProfile.Avatar
		model.claim = UserClaim{
			DisplayName: foundUserProfile.FullName,
			Email:       foundUserProfile.Email,
			UserId:      userAuth.ObjectId.String(),
			Role:        userAuth.Role,
			Avatar:      foundUserProfile.Avatar,
		}

	}

	return nil
}
func OauthGoogleCallback(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
	// Read oauthState from Cookie
	oauthState, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return handler.Response{
			Body:       []byte("Unauthorized OAuth callback."),
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	data, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return handler.Response{
			Body:       []byte("Unauthorized OAuth callback."),
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	// GetOrCreate User in your db.
	// Redirect or response with a token.
	// More code .....
	fmt.Fprintf(w, "UserInfo: %s\n", data)
	return handler.Response{
		Body:       []byte("Cookie generated."),
		StatusCode: http.StatusOK,
	}, nil
}

// OAuth2Handler makes a handler for OAuth 2.0 redirects
func OAuth2Handler(db interface{}) func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		config := &cf.AuthConfig
		c := &http.Client{
			Timeout: profileFetchTimeout,
		}

		clientSecret := config.ClientSecret

		if len(config.OAuthClientSecret) > 0 {
			clientSecret = strings.TrimSpace(config.OAuthClientSecret)
		}

		log.Printf(`OAuth 2 - "%s"`, r.URL.Path)
		if r.URL.Path != "/oauth2/authorized" {
			return handler.Response{
				Body:       []byte("Unauthorized OAuth callback."),
				StatusCode: http.StatusUnauthorized,
			}, nil
		}

		reqQuery := r.URL.Query()
		code := reqQuery.Get("code")
		state := reqQuery.Get("state")
		if len(code) == 0 {
			return handler.Response{
				Body:       []byte("Unauthorized OAuth callback, no code parameter given."),
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		if len(state) == 0 {
			return handler.Response{
				Body:       []byte("Unauthorized OAuth callback, no state parameter given."),
				StatusCode: http.StatusUnauthorized,
			}, nil
		}

		log.Printf("Exchange: %s, for an access_token", code)

		var tokenURL string
		var oauthProvider provider.Provider
		var redirectURI *url.URL

		switch config.OAuthProvider {
		case githubName:
			tokenURL = "https://github.com/login/oauth/access_token"
			oauthProvider = provider.NewGitHub(c)

			break
		case gitlabName:
			tokenURL = fmt.Sprintf("%s/oauth/token", config.OAuthProviderBaseURL)
			apiURL := config.OAuthProviderBaseURL + "/api/v4/"
			oauthProvider = provider.NewGitLabProvider(c, config.OAuthProviderBaseURL, apiURL)

			redirectAfterAutURL := reqQuery.Get("r")
			redirectURI, _ = url.Parse(combineURL(*coreConfig.AppConfig.Gateway, utils.GetPrettyURLf(config.BaseRoute+"/oauth2/authorized")))

			redirectURIQuery := redirectURI.Query()
			redirectURIQuery.Set("r", redirectAfterAutURL)

			redirectURI.RawQuery = redirectURIQuery.Encode()

			break
		}

		u, _ := url.Parse(tokenURL)
		q := u.Query()
		q.Set("client_id", config.ClientID)
		q.Set("client_secret", clientSecret)

		q.Set("code", code)
		q.Set("state", state)

		if config.OAuthProvider == gitlabName {
			q.Set("grant_type", "authorization_code")
			q.Set("redirect_uri", redirectURI.String())
		}

		u.RawQuery = q.Encode()
		log.Println("Posting to", u.String())

		newReq, _ := http.NewRequest(http.MethodPost, u.String(), nil)

		newReq.Header.Add("Accept", "application/json")
		res, err := c.Do(newReq)

		if err != nil {
			return handler.Response{
				Body:       []byte("Error exchanging code for access_token"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		token, tokenErr := getToken(res)
		if tokenErr != nil {
			log.Printf(
				"Unable to contact identity provider: %s, error: %s",
				config.OAuthProvider,
				tokenErr,
			)

			return handler.Response{
				Body: []byte(fmt.Sprintf(
					"Unable to contact identity provider: %s",
					config.OAuthProvider,
				)),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		fmt.Printf("\nGithub Token: %v\n", token.AccessToken)
		model := TokenModel{token: token, oauthProvider: oauthProvider, providerName: config.OAuthProvider}
		profile, profileErr := model.oauthProvider.GetProfile(model.token.AccessToken)
		if profileErr != nil {
			return handler.Response{
				Body: []byte(fmt.Sprintf(
					"Get oath profile error: %s",
					profileErr.Error(),
				)),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}
		model.profile = profile

		signupErr := checkSignup(token.AccessToken, &model, db)
		if signupErr != nil {
			log.Printf("Error signup: %s", signupErr.Error())
			return handler.Response{
				Body:       []byte("Internal server error signup check"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}
		session, err := createOAuthSession(&model)
		if err != nil {
			log.Printf("Error creating session: %s", err.Error())
			return handler.Response{
				Body:       []byte("Internal server error creating JWT"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		writeSessionOnCookie(w, session, config)

		log.Printf("SetCookie done, redirect to: %s", reqQuery)

		// Redirect to original requested resource (if specified in r=)
		redirect := reqQuery.Get("r")
		if len(redirect) > 0 {
			log.Printf(`Found redirect value "r"=%s, instructing client to redirect`, redirect)

			// Note: unable to redirect after setting Cookie, so landing on a redirect page instead.
			// http.Redirect(w, r, reqQuery.Get("r"), http.StatusTemporaryRedirect)

			return handler.Response{
				Body:       []byte(`<html><head></head>Redirecting.. <a href="redirect">to original resource</a>. <script>window.location.replace("` + redirect + `");</script></html>`),
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Content-Type": {" text/html; charset=utf-8"},
				},
			}, nil

		}

		webURL := utils.GetPrettyURLf("/web")

		return handler.Response{
			Body:       []byte(`<html><head></head>Redirecting.. <a href="redirect">to original resource</a>. <script>window.location.replace("` + webURL + `");</script></html>`),
			StatusCode: http.StatusOK,
			Header: map[string][]string{
				"Content-Type": {" text/html; charset=utf-8"},
			},
		}, nil
	}

}

func getUserOrganizations(username, accessToken string) (string, error) {

	organizations := []Organization{}
	apiURL := fmt.Sprintf("https://api.github.com/users/%s/orgs", username)

	req, reqErr := http.NewRequest(http.MethodGet, apiURL, nil)
	if reqErr != nil {
		return "", fmt.Errorf("error while making request to `%s` organizations: %s", apiURL, reqErr.Error())
	}
	req.Header.Add("Authorization", "token "+accessToken)

	client := http.DefaultClient
	resp, respErr := client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if respErr != nil {
		return "", fmt.Errorf("error while requesting organizations: %s", respErr.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code from request to GitHub organizations: %d", resp.StatusCode)
	}

	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return "", fmt.Errorf("error while reading body from GitHub organizations: %s", bodyErr.Error())
	}

	var allOrganizations []string
	unmarshallErr := json.Unmarshal(body, &organizations)
	if unmarshallErr != nil {
		return "", fmt.Errorf("error while un-marshaling organizations: %s", unmarshallErr.Error())
	}

	for _, organization := range organizations {
		allOrganizations = append(allOrganizations, organization.Login)
	}
	formatOrganizations := strings.Join(allOrganizations, ",")

	return formatOrganizations, nil
}

type Organization struct {
	Login string `json:"login"`
}
