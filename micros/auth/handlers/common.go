package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alexellis/hmac"
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/gofrs/uuid"
	af "github.com/red-gold/telar-core/config"
	coreConfig "github.com/red-gold/telar-core/config"
	tsconfig "github.com/red-gold/telar-core/config"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	cf "github.com/red-gold/telar-web/micros/auth/config"
	"github.com/red-gold/telar-web/micros/auth/provider"
	settingModels "github.com/red-gold/telar-web/micros/setting/models"
)

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

// ProviderAccessToken as issued by GitHub or GitLab
type ProviderAccessToken struct {
	AccessToken string `json:"access_token"`
}

func buildGitHubURL(config *cf.Configuration, string, scope string) *url.URL {
	authURL := "https://github.com/login/oauth/authorize"
	u, _ := url.Parse(authURL)
	q := u.Query()

	q.Set("scope", scope)
	q.Set("allow_signup", "0")
	q.Set("state", fmt.Sprintf("%d", time.Now().Unix()))
	q.Set("client_id", config.ClientID)

	redirectURI := combineURL(*coreConfig.AppConfig.Gateway, utils.GetPrettyURLf(config.BaseRoute+"/oauth2/authorized"))

	q.Set("redirect_uri", redirectURI)

	u.RawQuery = q.Encode()
	return u
}

func buildGitLabURL(config *cf.Configuration) *url.URL {
	authURL := config.OAuthProviderBaseURL + "/oauth/authorize"

	u, _ := url.Parse(authURL)
	q := u.Query()

	q.Set("client_id", config.ClientID)
	q.Set("response_type", "code")
	q.Set("state", fmt.Sprintf("%d", time.Now().Unix()))

	redirectURI := combineURL(*coreConfig.AppConfig.Gateway, utils.GetPrettyURLf(config.BaseRoute+"/oauth2/authorized"))

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
func writeSessionOnCookie(w http.ResponseWriter, session string, config *cf.Configuration) {
	appConfig := af.AppConfig
	parts := strings.Split(session, ".")
	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Name:     *appConfig.HeaderCookieName,
		Value:    parts[0],
		Path:     "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	})

	http.SetCookie(w, &http.Cookie{
		// HttpOnly: true,
		Name:  *appConfig.PayloadCookieName,
		Value: parts[1],
		Path:  "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	})

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Name:     *appConfig.SignatureCookieName,
		Value:    parts[2],
		Path:     "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	})
}

// createToken
func createToken(model *TokenModel) (string, error) {
	var err error
	var session string
	authConfig := &cf.AuthConfig
	coreConfig := &tsconfig.AppConfig

	privateKey, keyErr := jwt.ParseECPrivateKeyFromPEM([]byte(*coreConfig.PrivateKey))
	if keyErr != nil {
		log.Fatalf("unable to parse private key: %s", keyErr.Error())
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodES256.Name)
	claims := TelarSocailClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        fmt.Sprintf("%s", model.profile.ID),
			Issuer:    fmt.Sprintf("telar-social@%s", model.providerName),
			ExpiresAt: time.Now().Add(48 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   model.profile.Login,
			Audience:  authConfig.CookieRootDomain,
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

// emailCodeVerifyTmpl
func emailCodeVerifyTmpl(username string, code string, appName string) (string, error) {
	return utils.ParseHtmlTemplate("html_template/email_code_verify_reset_pass.html",
		htmlTemplate{Name: username, Code: code, AppName: appName})
}

// emailCodeVerifyResetPassTmpl
func emailCodeVerifyResetPassTmpl(username string, code string, appName string, email string) (string, error) {
	return utils.ParseHtmlTemplate("html_template/email_code_verify_reset_pass.html",
		htmlTemplate{Name: username, Code: code, AppName: appName, Email: email})
}

func phoneVerifyCode(code string, appName string) string {
	return fmt.Sprintf("Verfy code from %s : %s", code, appName)
}

// functionCallByCookie send request to another function/microservice using cookie validation
func functionCallByHeader(method string, bytesReq []byte, url string, header map[string][]string) ([]byte, error) {
	prettyURL := utils.GetPrettyURLf(url)
	bodyReader := bytes.NewBuffer(bytesReq)

	httpReq, httpErr := http.NewRequest(method, *coreConfig.AppConfig.Gateway+prettyURL, bodyReader)
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
		return nil, fmt.Errorf("failed to call %s api, invalid status: %s", prettyURL, res.Status)
	}

	return resData, nil
}

func initUserSetup(userId uuid.UUID, email string, avatar string, displayName string, role string) error {

	// Create admin header for http request
	adminHeaders := make(map[string][]string)
	adminHeaders["uid"] = []string{userId.String()}
	adminHeaders["email"] = []string{email}
	adminHeaders["avatar"] = []string{avatar}
	adminHeaders["displayName"] = []string{displayName}
	adminHeaders["role"] = []string{role}

	circleURL := fmt.Sprintf("/circles/following/%s", userId)
	_, circleErr := functionCall([]byte(""), circleURL)

	if circleErr != nil {
		return circleErr
	}

	// Create default setting for user
	settingModel := settingModels.CreateSettingGroupModel{
		Type: "notification",
		List: []settingModels.SettingGroupItemModel{
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
	}

	settingBytes, marshalErr := json.Marshal(&settingModel)
	if marshalErr != nil {
		return marshalErr
	}

	// Send request for setting
	settingURL := "/setting"
	_, settingErr := functionCallByHeader(http.MethodPost, settingBytes, settingURL, adminHeaders)

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
	_, actionRoomErr := functionCallByHeader(http.MethodPost, actiomRoomBytes, actionRoomURL, adminHeaders)

	if actionRoomErr != nil {
		return actionRoomErr
	}

	return nil
}

func generateRandomNumber(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return (rand.Intn(max-min+1) + min)
}
