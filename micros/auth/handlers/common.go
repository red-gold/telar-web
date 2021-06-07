package handlers

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alexellis/hmac"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	coreConfig "github.com/red-gold/telar-core/config"
	log "github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	authConfig "github.com/red-gold/telar-web/micros/auth/config"
	"github.com/red-gold/telar-web/micros/auth/models"
	"github.com/red-gold/telar-web/micros/auth/provider"
)

type UserInfoInReq struct {
	UserId      uuid.UUID `json:"userId"`
	Username    string    `json:"username"`
	Avatar      string    `json:"avatar"`
	DisplayName string    `json:"displayName"`
	SystemRole  string    `json:"systemRole"`
}

type htmlTemplate struct {
	Name    string
	Code    string
	AppName string
	Email   string
}

// User information claim
type UserClaim struct {
	DisplayName   string `json:"displayName"`
	Organizations string `json:"organizations"`
	Avatar        string `json:"avatar"`
	UserId        string `json:"uid"`
	Email         string `json:"email"`
	Role          string `json:"role"`
}

type TokenModel struct {
	token            ProviderAccessToken
	oauthProvider    provider.Provider
	providerName     string
	organizationList string
	profile          *provider.Profile
	claim            UserClaim
}

type CreateActionRoomModel struct {
	ObjectId    uuid.UUID `json:"objectId"`
	OwnerUserId uuid.UUID `json:"ownerUserId"`
	PrivateKey  string    `json:"privateKey"`
	AccessKey   string    `json:"accessKey"`
	Status      int       `json:"status"`
	CreatedDate int64     `json:"created_date"`
}

// TelarSocailClaims extends standard claims
type TelarSocailClaims struct {
	// Name is the full name of the user for OIDC
	Name string `json:"name"`

	// AccessToken for use with the GitHub Profile API
	AccessToken string `json:"access_token"`

	// String with all organizations separated with commas
	Organizations string `json:"organizations"`

	// User information
	Claim UserClaim `json:"claim"`

	// Inherit from standard claims
	jwt.StandardClaims
}

type ResetPasswordClaims struct {
	VerifyId string `json:"verifyId"`
	jwt.StandardClaims
}

// ProviderAccessToken as issued by GitHub or GitLab
type ProviderAccessToken struct {
	AccessToken string `json:"access_token"`
}

type ProfileResultAsync struct {
	Profile *models.UserProfileModel
	Error   error
}

type UsersLangSettingsResultAsync struct {
	settings map[string]string
	Error    error
}

// getHeadersFromUserInfoReq
func getHeadersFromUserInfoReq(info *UserInfoInReq) map[string][]string {
	userHeaders := make(map[string][]string)
	userHeaders["uid"] = []string{info.UserId.String()}
	userHeaders["email"] = []string{info.Username}
	userHeaders["avatar"] = []string{info.Avatar}
	userHeaders["displayName"] = []string{info.DisplayName}
	userHeaders["role"] = []string{info.SystemRole}

	return userHeaders
}

// getUserInfoReq
func getUserInfoReq(c *fiber.Ctx) *UserInfoInReq {
	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		return &UserInfoInReq{}
	}
	userInfoInReq := &UserInfoInReq{
		UserId:      currentUser.UserID,
		Username:    currentUser.Username,
		Avatar:      currentUser.Avatar,
		DisplayName: currentUser.DisplayName,
		SystemRole:  currentUser.SystemRole,
	}
	return userInfoInReq

}

// getSettingPath
func getSettingPath(userId uuid.UUID, settingType, settingKey string) string {
	key := fmt.Sprintf("%s:%s:%s", userId.String(), settingType, settingKey)
	return key
}

// decodeResetPasswordToken Decode reset password token
func decodeResetPasswordToken(token string) (*ResetPasswordClaims, error) {

	// Create the JWT key used to decode token
	privateKeyEnc := b64.StdEncoding.EncodeToString([]byte(*coreConfig.AppConfig.PrivateKey))
	var jwtKey = []byte(privateKeyEnc[:20])

	decodedToken, err := b64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("Can not decode token %s", err.Error())
	}

	claims := new(ResetPasswordClaims)

	tkn, err := jwt.ParseWithClaims(string(decodedToken), claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, fmt.Errorf("signatureInvalid")
		}
		return nil, err
	}
	if !tkn.Valid {
		return nil, fmt.Errorf("invalidToken")
	}
	return claims, nil
}

// generateResetPasswordToken Generate reset password token
func generateResetPasswordToken(verifyId string) (string, error) {

	// Create the JWT key used to create the signature
	privateKeyEnc := b64.StdEncoding.EncodeToString([]byte(*coreConfig.AppConfig.PrivateKey))
	var jwtKey = []byte(privateKeyEnc[:20])

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(5 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &ResetPasswordClaims{
		VerifyId: verifyId,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error\
		return "", err
	}

	return b64.StdEncoding.EncodeToString([]byte(tokenString)), nil

}

func buildGitHubURL(config *authConfig.Configuration, string, scope string) *url.URL {
	authURL := "https://github.com/login/oauth/authorize"
	u, _ := url.Parse(authURL)
	q := u.Query()

	q.Set("scope", scope)
	q.Set("allow_signup", "0")
	q.Set("state", fmt.Sprintf("%d", time.Now().Unix()))
	q.Set("client_id", config.ClientID)

	redirectURI := combineURL(config.AuthWebURI, utils.GetPrettyURLf(config.BaseRoute+"/oauth2/authorized"))

	q.Set("redirect_uri", redirectURI)

	u.RawQuery = q.Encode()
	return u
}

func buildGitLabURL(config *authConfig.Configuration) *url.URL {
	authURL := config.OAuthProviderBaseURL + "/oauth/authorize"

	u, _ := url.Parse(authURL)
	q := u.Query()

	q.Set("client_id", config.ClientID)
	q.Set("response_type", "code")
	q.Set("state", fmt.Sprintf("%d", time.Now().Unix()))

	redirectURI := combineURL(config.AuthWebURI, utils.GetPrettyURLf(combineURL(config.BaseRoute, "/oauth2/authorized")))

	q.Set("redirect_uri", redirectURI)

	u.RawQuery = q.Encode()

	return u
}

func combineURL(a, b string) string {
	if !strings.HasSuffix(a, "/") {
		a = a + "/"
	}
	if strings.HasPrefix(b, "/") {
		b = strings.TrimPrefix(b, "/")
	}

	return a + b
}

func createOAuthSession(model *TokenModel) (string, error) {
	fmt.Printf("\nToken Model: %v\n", model)

	model.organizationList = ""

	if model.providerName == "github" {
		fmt.Printf("\nGithub Access Token: %s\n", model.token.AccessToken)
		organizations, organizationsErr := getUserOrganizations(model.profile.Login, model.token.AccessToken)
		if organizationsErr != nil {
			return "", organizationsErr
		}
		model.organizationList = organizations
	}
	return createToken(model)
}

// writeTokenOnCookie wite session on cookie
func writeSessionOnCookie(c *fiber.Ctx, session string, config *authConfig.Configuration) {
	appConfig := coreConfig.AppConfig
	parts := strings.Split(session, ".")
	headerCookie := &fiber.Cookie{
		HTTPOnly: true,
		Name:     *appConfig.HeaderCookieName,
		Value:    parts[0],
		Path:     "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	}

	payloadCookie := &fiber.Cookie{
		// HttpOnly: true,
		Name:  *appConfig.PayloadCookieName,
		Value: parts[1],
		Path:  "/",
		// Expires: time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	}

	signCookie := &fiber.Cookie{
		HTTPOnly: true,
		Name:     *appConfig.SignatureCookieName,
		Value:    parts[2],
		Path:     "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	}
	// Set cookie
	c.Cookie(headerCookie)
	c.Cookie(payloadCookie)
	c.Cookie(signCookie)
}

// Write user language on cookie
func writeUserLangOnCookie(c *fiber.Ctx, lang string) {
	langCookie := &fiber.Cookie{
		HTTPOnly: false,
		Name:     "social-lang",
		Value:    lang,
		Path:     "/",
		Domain:   authConfig.AuthConfig.CookieRootDomain,
	}
	c.Cookie(langCookie)
}

// createToken
func createToken(model *TokenModel) (string, error) {
	var err error
	var session string

	privateKey, keyErr := jwt.ParseECPrivateKeyFromPEM([]byte(*coreConfig.AppConfig.PrivateKey))
	if keyErr != nil {
		log.Error("unable to parse private key: %s", keyErr.Error())
		return "", fmt.Errorf("unable to parse private key: %s", keyErr.Error())
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodES256.Name)
	claims := TelarSocailClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        fmt.Sprintf("%s", model.profile.ID),
			Issuer:    fmt.Sprintf("telar-social@%s", model.providerName),
			ExpiresAt: time.Now().Add(48 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   model.profile.Login,
			Audience:  authConfig.AuthConfig.CookieRootDomain,
		},
		Organizations: model.organizationList,
		Name:          model.profile.Name,
		AccessToken:   model.token.AccessToken,
		Claim:         model.claim,
	}

	session, err = jwt.NewWithClaims(method, claims).SignedString(privateKey)

	return session, err
}

// getToken
func getToken(res *http.Response) (ProviderAccessToken, error) {
	token := ProviderAccessToken{}
	if res.Body != nil {
		defer res.Body.Close()

		tokenRes, _ := ioutil.ReadAll(res.Body)

		err := json.Unmarshal(tokenRes, &token)
		if err != nil {
			return token, err
		}
		return token, nil
	}

	return token, fmt.Errorf("no body received from server")
}

func phoneVerifyCode(code string, appName string) string {
	return fmt.Sprintf("Verfy code from %s : %s", code, appName)
}

// functionCall send request to another function/microservice using cookie validation
func functionCall(method string, bytesReq []byte, url string, header map[string][]string) ([]byte, error) {
	prettyURL := utils.GetPrettyURLf(url)
	bodyReader := bytes.NewBuffer(bytesReq)

	httpReq, httpErr := http.NewRequest(method, *coreConfig.AppConfig.InternalGateway+prettyURL, bodyReader)
	if httpErr != nil {
		return nil, httpErr
	}
	payloadSecret := *coreConfig.AppConfig.PayloadSecret

	digest := hmac.Sign(bytesReq, []byte(payloadSecret))
	httpReq.Header.Set("Content-type", "application/json")
	fmt.Printf("\ndigest: %s, header: %v \n", "sha1="+hex.EncodeToString(digest), types.HeaderHMACAuthenticate)
	httpReq.Header.Add(types.HeaderHMACAuthenticate, "sha1="+hex.EncodeToString(digest))
	if header != nil {
		for k, v := range header {
			httpReq.Header[k] = v
		}
	}
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
		if res.StatusCode == http.StatusNotFound {
			return nil, NotFoundHTTPStatusError
		}
		return nil, fmt.Errorf("failed to call %s api, invalid status: %s", prettyURL, res.Status)
	}

	return resData, nil
}

// createDefaultLangSetting
func createDefaultLangSetting(userInfoInReq *UserInfoInReq) error {

	settingModel := models.CreateMultipleSettingsModel{
		List: []models.CreateSettingGroupModel{
			{
				Type: "lang",
				List: []models.SettingGroupItemModel{
					{
						Name:  "current",
						Value: "en",
					},
				},
			},
		},
	}

	settingBytes, marshalErr := json.Marshal(&settingModel)
	if marshalErr != nil {
		return marshalErr
	}

	// Send request for setting
	settingURL := "/setting"
	_, settingErr := functionCall(http.MethodPost, settingBytes, settingURL, getHeadersFromUserInfoReq(userInfoInReq))

	if settingErr != nil {
		return settingErr
	}
	return nil
}

// initUserSetup
func initUserSetup(userId uuid.UUID, email string, avatar string, displayName string, role string) error {

	// Create admin header for http request
	adminHeaders := make(map[string][]string)
	adminHeaders["uid"] = []string{userId.String()}
	adminHeaders["email"] = []string{email}
	adminHeaders["avatar"] = []string{avatar}
	adminHeaders["displayName"] = []string{displayName}
	adminHeaders["role"] = []string{role}

	circleURL := fmt.Sprintf("/circles/following/%s", userId)
	_, circleErr := functionCall(http.MethodPost, []byte(""), circleURL, adminHeaders)

	if circleErr != nil {
		return circleErr
	}

	// Create default setting for user
	settingModel := models.CreateMultipleSettingsModel{
		List: []models.CreateSettingGroupModel{
			{
				Type: "notification",
				List: []models.SettingGroupItemModel{
					{
						Name:  "send_email_on_like",
						Value: "false",
					},
					{
						Name:  "send_email_on_follow",
						Value: "false",
					},
					{
						Name:  "send_email_on_comment_post",
						Value: "false",
					},
				},
			},
			{
				Type: "lang",
				List: []models.SettingGroupItemModel{
					{
						Name:  "current",
						Value: "en",
					},
				},
			},
		},
	}

	settingBytes, marshalErr := json.Marshal(&settingModel)
	if marshalErr != nil {
		return marshalErr
	}

	// Send request for setting
	settingURL := "/setting"
	_, settingErr := functionCall(http.MethodPost, settingBytes, settingURL, adminHeaders)

	if settingErr != nil {
		return settingErr
	}

	privateKey, privateKeyErr := uuid.NewV4()
	if privateKeyErr != nil {
		return privateKeyErr
	}

	accessKey, accessKeyErr := uuid.NewV4()
	if accessKeyErr != nil {
		return accessKeyErr
	}

	// Send request for action room
	actionRoomModel := CreateActionRoomModel{
		OwnerUserId: userId,
		PrivateKey:  privateKey.String(),
		AccessKey:   accessKey.String(),
	}

	actiomRoomBytes, marshalErr := json.Marshal(&actionRoomModel)
	if marshalErr != nil {
		return marshalErr
	}
	actionRoomURL := "/actions/room"
	_, actionRoomErr := functionCall(http.MethodPost, actiomRoomBytes, actionRoomURL, adminHeaders)

	if actionRoomErr != nil {
		return actionRoomErr
	}

	return nil
}

// getUserProfileByID Get user profile by user ID
func getUserProfileByID(userID uuid.UUID) (*models.UserProfileModel, error) {
	profileURL := fmt.Sprintf("/profile/dto/id/%s", userID.String())
	foundProfileData, err := functionCall(http.MethodGet, []byte(""), profileURL, nil)
	if err != nil {
		if err == NotFoundHTTPStatusError {
			return nil, nil
		}
		log.Error("functionCall (%s) -  %s", profileURL, err.Error())
		return nil, fmt.Errorf("getUserProfileByID/functionCall")
	}
	var foundProfile models.UserProfileModel
	err = json.Unmarshal(foundProfileData, &foundProfile)
	if err != nil {
		log.Error("Unmarshal foundProfile -  %s", err.Error())
		return nil, fmt.Errorf("getUserProfileByID/unmarshal")
	}
	return &foundProfile, nil
}

// saveUserProfile Save user profile
func saveUserProfile(model *models.UserProfileModel) error {
	profileURL := "/profile/dto"
	data, err := json.Marshal(model)
	if err != nil {
		log.Error("marshal models.UserProfileModel %s", err.Error())
		return fmt.Errorf("saveProfile/marshalUserProfileModel")
	}
	_, err = functionCall(http.MethodPost, data, profileURL, nil)
	if err != nil {
		log.Error("functionCall (%s) -  %s", profileURL, err.Error())
		return fmt.Errorf("saveUserProfile/functionCall")
	}
	return nil
}

// updateUserProfile Update user profile
func updateUserProfile(model *models.ProfileUpdateModel, userId uuid.UUID, email, avatar, displayName, role string) error {
	profileURL := "/profile"
	data, err := json.Marshal(model)
	if err != nil {
		log.Error("marshal models.UserProfileModel %s", err.Error())
		return fmt.Errorf("saveProfile/marshalUserProfileModel")
	}
	headers := make(map[string][]string)
	headers["uid"] = []string{userId.String()}
	headers["email"] = []string{email}
	headers["avatar"] = []string{avatar}
	headers["displayName"] = []string{displayName}
	headers["role"] = []string{role}
	_, err = functionCall(http.MethodPut, data, profileURL, headers)
	if err != nil {
		log.Error("functionCall (%s) -  %s", profileURL, err.Error())
		return fmt.Errorf("updateUserProfile/functionCall")
	}
	return nil
}

// generateRandomNumber
func generateRandomNumber(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return (rand.Intn(max-min+1) + min)
}

// readProfileAsync Read profile async
func readProfileAsync(userID uuid.UUID) <-chan ProfileResultAsync {
	r := make(chan ProfileResultAsync)
	go func() {
		defer close(r)

		profile, err := getUserProfileByID(userID)
		if err != nil {
			r <- ProfileResultAsync{Error: err}
			return
		}
		r <- ProfileResultAsync{Profile: profile}

	}()
	return r
}

// getUsersLangSettings Get users language settings
func getUsersLangSettings(userIds []uuid.UUID, userInfoInReq *UserInfoInReq) (map[string]string, error) {
	url := "/setting/dto/ids"
	model := models.GetSettingsModel{
		UserIds: userIds,
		Type:    "lang",
	}
	payload, marshalErr := json.Marshal(model)
	if marshalErr != nil {
		return nil, marshalErr
	}

	resData, callErr := functionCall(http.MethodPost, payload, url, getHeadersFromUserInfoReq(userInfoInReq))
	if callErr != nil {

		return nil, fmt.Errorf("Cannot send request to %s - %s", url, callErr.Error())
	}

	var parsedData map[string]string
	json.Unmarshal(resData, &parsedData)
	return parsedData, nil
}

// readLanguageSettingAsync Read language setting async
func readLanguageSettingAsync(userID uuid.UUID, userInfoInReq *UserInfoInReq) <-chan UsersLangSettingsResultAsync {
	r := make(chan UsersLangSettingsResultAsync)
	go func() {
		defer close(r)

		settings, err := getUsersLangSettings([]uuid.UUID{userID}, userInfoInReq)
		if err != nil {
			r <- UsersLangSettingsResultAsync{Error: err}
			return
		}
		r <- UsersLangSettingsResultAsync{settings: settings}

	}()
	return r
}
