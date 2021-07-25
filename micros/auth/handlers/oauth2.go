package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/utils"
	"github.com/red-gold/telar-web/constants"
	cf "github.com/red-gold/telar-web/micros/auth/config"
	"github.com/red-gold/telar-web/micros/auth/database"
	dto "github.com/red-gold/telar-web/micros/auth/dto"
	"github.com/red-gold/telar-web/micros/auth/models"
	"github.com/red-gold/telar-web/micros/auth/provider"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

const profileFetchTimeout = time.Second * 5

// checkOAuthSignup check for user oauth signup in the case user does not exist in user auth
func checkOAuthSignup(accessToken string, model *TokenModel, currentUserLang *string, db interface{}) error {

	if model.profile.Name == "" {
		log.Error("[ERROR]: OAuth provide - name can not be empty")
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

	fmt.Printf("[INFO]: Oauth check signup for user %s", model.profile.Email)

	// Check user exist
	userAuth, findError := userAuthService.FindByUsername(model.profile.Email)
	if findError != nil {
		errorMessage := fmt.Sprintf("[ERROR]: Error while finding user by user name : %s",
			findError.Error())
		fmt.Println(errorMessage)

	}
	fmt.Printf("[INFO]: Oauth check signup - user auth object %v", userAuth)

	if userAuth == nil {
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
		newUserProfile := &models.UserProfileModel{
			ObjectId:    newUserId,
			FullName:    model.profile.Name,
			SocialName:  generateSocialName(model.profile.Name, newUserId.String()),
			CreatedDate: createdDate,
			LastUpdated: createdDate,
			Email:       model.profile.Email,
			Avatar:      model.profile.Avatar,
			Banner:      fmt.Sprintf("https://picsum.photos/id/%d/900/300/?blur", generateRandomNumber(1, 1000)),
			Permission:  constants.Public,
		}
		userProfileErr := saveUserProfile(newUserProfile)
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
			Banner:      newUserProfile.Banner,
			TagLine:     newUserProfile.TagLine,
			CreatedDate: newUserProfile.CreatedDate,
		}
	} else {

		profileChannel := readProfileAsync(userAuth.ObjectId)
		langChannel := readLanguageSettingAsync(userAuth.ObjectId, &UserInfoInReq{
			UserId:      userAuth.ObjectId,
			Username:    userAuth.Username,
			Avatar:      "",
			DisplayName: "",
			SystemRole:  userAuth.Role,
		})

		profileResult, langResult := <-profileChannel, <-langChannel
		if profileResult.Error != nil || profileResult.Profile == nil {
			if profileResult.Error != nil {
				fmt.Printf("\n User profile  %s\n", profileResult.Error.Error())
			}
			fmt.Println("\n Could not find user profile", userAuth.ObjectId)
			return fmt.Errorf("Could not find user profile %s", userAuth.ObjectId)
		}

		*currentUserLang = "en"
		langSettigPath := getSettingPath(userAuth.ObjectId, "lang", "current")
		if val, ok := langResult.settings[langSettigPath]; ok && val != "" {
			*currentUserLang = val
		} else {
			go func() {
				userInfoReq := &UserInfoInReq{
					UserId:      userAuth.ObjectId,
					Username:    userAuth.Username,
					Avatar:      profileResult.Profile.Avatar,
					DisplayName: profileResult.Profile.FullName,
					SystemRole:  userAuth.Role,
				}
				createDefaultLangSetting(userInfoReq)
			}()
		}

		model.profile.ID = userAuth.ObjectId.String()
		model.profile.Email = profileResult.Profile.Email
		model.profile.Name = profileResult.Profile.FullName
		model.profile.Avatar = profileResult.Profile.Avatar
		model.claim = UserClaim{
			DisplayName: profileResult.Profile.FullName,
			Email:       profileResult.Profile.Email,
			UserId:      userAuth.ObjectId.String(),
			Role:        userAuth.Role,
			Avatar:      profileResult.Profile.Avatar,
			Banner:      profileResult.Profile.Banner,
			TagLine:     profileResult.Profile.TagLine,
			CreatedDate: userAuth.CreatedDate,
		}

	}

	return nil
}

// OAuth2Handler makes a handler for OAuth 2.0 redirects
func OAuth2Handler(c *fiber.Ctx) error {

	config := &cf.AuthConfig
	httpClient := &http.Client{
		Timeout: profileFetchTimeout,
	}

	clientSecret := config.ClientSecret

	if len(config.OAuthClientSecret) > 0 {
		clientSecret = strings.TrimSpace(config.OAuthClientSecret)
	}

	log.Info(`OAuth 2 - "%s"`, c.Path())
	if c.Path() != "/oauth2/authorized" {
		return c.Status(http.StatusUnauthorized).JSON(utils.Error("internal/oAuthPath", "Unauthorized OAuth callback."))
	}

	code := c.Query("code")
	state := c.Query("state")
	if len(code) == 0 {
		return c.Status(http.StatusUnauthorized).JSON(utils.Error("internal/oAuthCode", "Unauthorized OAuth callback, no code parameter given!"))
	}
	if len(state) == 0 {
		return c.Status(http.StatusUnauthorized).JSON(utils.Error("internal/oAuthState", "Unauthorized OAuth callback, no state parameter given!"))
	}

	log.Info("Exchange: %s, for an access_token", code)

	var tokenURL string
	var oauthProvider provider.Provider
	var redirectURI *url.URL

	switch config.OAuthProvider {
	case githubName:
		tokenURL = "https://github.com/login/oauth/access_token"
		oauthProvider = provider.NewGitHub(httpClient)

		break
	case gitlabName:
		tokenURL = fmt.Sprintf("%s/oauth/token", config.OAuthProviderBaseURL)
		apiURL := config.OAuthProviderBaseURL + "/api/v4/"
		oauthProvider = provider.NewGitLabProvider(httpClient, config.OAuthProviderBaseURL, apiURL)

		redirectAfterAutURL := c.Query("r")
		redirectURI, _ = url.Parse(combineURL(config.AuthWebURI, utils.GetPrettyURLf(config.BaseRoute+"/oauth2/authorized")))

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
	log.Info("Posting to %s", u.String())

	newReq, _ := http.NewRequest(http.MethodPost, u.String(), nil)

	newReq.Header.Add("Accept", "application/json")
	res, err := httpClient.Do(newReq)

	if err != nil {
		log.Error("Error exchanging code for access_token %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/identityProvider", "Error exchanging code for access_token!"))
	}

	token, tokenErr := getToken(res)
	if tokenErr != nil {
		log.Error(
			"Unable to contact identity provider: %s, error: %s",
			config.OAuthProvider,
			tokenErr,
		)
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/identityProvider", "Unable to contact identity provider!"))
	}

	fmt.Printf("\nGithub Token: %v\n", token.AccessToken)
	model := TokenModel{token: token, oauthProvider: oauthProvider, providerName: config.OAuthProvider}
	profile, profileErr := model.oauthProvider.GetProfile(model.token.AccessToken)
	if profileErr != nil {
		log.Error("Get OAuth profile %s", profileErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/getOAuthProfile", "Get oath profile error!"))
	}
	model.profile = profile
	var currentUserLang string
	signupErr := checkOAuthSignup(token.AccessToken, &model, &currentUserLang, database.Db)
	if signupErr != nil {
		log.Error("Error signup: %s", signupErr.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/signupCheck", "Internal server error signup check!"))
	}

	session, err := createOAuthSession(&model)
	if err != nil {
		log.Error("Error creating session: %s", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(utils.Error("internal/createToken", "Internal server error creating token!"))
	}

	writeSessionOnCookie(c, session, config)

	// Write user language on cookie
	writeUserLangOnCookie(c, currentUserLang)

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

	webURL := config.ExternalRedirectDomain

	return c.Render("redirect", fiber.Map{
		"URL": webURL,
	})

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
