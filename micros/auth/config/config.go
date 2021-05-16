package config

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	coreUtils "github.com/red-gold/telar-core/utils"
)

const (
	basePath               = "/var/openfaas/secrets/"
	oauthClientSecretKey   = "ts-client-secret"
	adminUsernameSecretKey = "admin-username"
	adminPasswordSecretKey = "admin-password"
)

var secretKeys = []string{oauthClientSecretKey}

// Initialize AppConfig
func InitConfig() {

	// Load config from environment values if exists
	loadAllConfig()
}

// getAllConfigFromFile get all config from files
func getAllConfigFromFile() map[string][]byte {
	filePaths := []string{}
	for _, v := range secretKeys {
		filePaths = append(filePaths, basePath+v)
	}
	return coreUtils.GetFilesContents(filePaths...)
}

// Load config from environment
func loadAllConfig() {

	loadSecretMode, ok := os.LookupEnv("load_secret_mode")
	if ok {
		log.Printf("[INFO]: Load secret mode information loaded from env.")
		if loadSecretMode == "env" {
			loadSecretsFromEnv()
		}
	} else {
		log.Printf("[INFO]: No secret mode in env. Secrets are loading from file.")
		loadSecretsFromFile()
	}

	oauthProvider, ok := os.LookupEnv("oauth_provider")
	if ok {
		AuthConfig.OAuthProvider = oauthProvider
		log.Printf("[INFO]: OAuthProvider information loaded from env.")
	}

	oauthProviderBaseUrl, ok := os.LookupEnv("oauth_provider_base_url")
	if ok {
		AuthConfig.OAuthProviderBaseURL = oauthProviderBaseUrl
		log.Printf("[INFO]: OAuthProviderBaseURL information loaded from env.")
	}

	oauthTelarBaseUrl, ok := os.LookupEnv("oauth_telar_base_url")
	if ok {
		AuthConfig.OAuthTelarBaseURL = oauthTelarBaseUrl
		log.Printf("[INFO]: OAuthTelarBaseURL information loaded from env.")
	}

	clientId, ok := os.LookupEnv("client_id")
	if ok {
		AuthConfig.ClientID = clientId
		log.Printf("[INFO]: ClientID information loaded from env.")
	}

	clientSecret, ok := os.LookupEnv("client_secret")
	if ok {
		AuthConfig.ClientSecret = clientSecret
		log.Printf("[INFO]: ClientSecret information loaded from env.")
	}

	externalRedirectDomain, ok := os.LookupEnv("external_redirect_domain")
	if ok {
		AuthConfig.ExternalRedirectDomain = externalRedirectDomain
		log.Printf("[INFO]: ExternalRedirectDomain information loaded from env.")
	}

	authWebURI, ok := os.LookupEnv("auth_web_uri")
	if ok {
		AuthConfig.AuthWebURI = authWebURI
		log.Printf("[INFO]: AuthWebURI information loaded from env [%s] ", authWebURI)
	}

	WebURL, ok := os.LookupEnv("web_url")
	if ok {
		AuthConfig.WebURL = WebURL
		log.Printf("[INFO]: Web URL information loaded from env [%s] ", WebURL)
	}

	scope, ok := os.LookupEnv("oauth_scope")
	if ok {
		AuthConfig.Scope = scope
		log.Printf("[INFO]: OAuth Scope information loaded from env.")
	}

	cookieRootDomain, ok := os.LookupEnv("cookie_root_domain")
	if ok {
		AuthConfig.CookieRootDomain = cookieRootDomain
		log.Printf("[INFO]: CookieRootDomain information loaded from env.")
	}

	cookieExpiry, ok := os.LookupEnv("cookie_expiry")
	if ok {
		expireTime, atoiErr := strconv.Atoi(cookieExpiry)
		if atoiErr != nil {
			log.Printf("[Error]: Information loding from got error: %s.", atoiErr.Error())
		} else {
			ext := time.Hour * time.Duration(expireTime)
			AuthConfig.CookieExpiresIn = ext
			log.Printf("[INFO]: CookieExpiresIn information loaded from env.")
		}
	}

	baseRoute, ok := os.LookupEnv("base_route")
	if ok {
		AuthConfig.BaseRoute = baseRoute
		log.Printf("[INFO]: Base route information loaded from env.")
	}

	verifyType, ok := os.LookupEnv("verify_type")
	if ok {
		AuthConfig.VerifyType = verifyType
		log.Printf("[INFO]: Base route information loaded from env.")
	}

	debug, ok := os.LookupEnv("write_debug")
	if ok {
		parsedDebug, errParseDebug := strconv.ParseBool(debug)
		if errParseDebug != nil {
			log.Printf("[ERROR]: Debug information loading error: %s", errParseDebug.Error())
		}
		AuthConfig.Debug = parsedDebug
		log.Printf("[INFO]: Debug information loaded from env.")
	}
}

// loadSecretsFromFile Load secrets from file
func loadSecretsFromFile() {
	filesConfig := getAllConfigFromFile()

	if filesConfig[basePath+oauthClientSecretKey] != nil {
		oauthClientSecret := string(filesConfig[basePath+oauthClientSecretKey])
		AuthConfig.OAuthClientSecret = oauthClientSecret
		log.Printf("[INFO]: OAuth client secret information loaded from env.")
	}

	if filesConfig[basePath+adminUsernameSecretKey] != nil {
		adminUsername := string(filesConfig[basePath+adminUsernameSecretKey])
		AuthConfig.AdminUsername = adminUsername
		log.Printf("[INFO]: Admin username information loaded from env.")
	}

	if filesConfig[basePath+adminPasswordSecretKey] != nil {
		adminPassword := string(filesConfig[basePath+adminPasswordSecretKey])
		AuthConfig.AdminPassword = adminPassword
		log.Printf("[INFO]: Admin password information loaded from env.")
	}

}

// loadSecretsFromEnv Load secrets from environment variables
func loadSecretsFromEnv() {

	oauthClientSecret, ok := os.LookupEnv("ts_client_secret")
	if ok {
		oauthClientSecret = decodeBase64(oauthClientSecret)
		AuthConfig.OAuthClientSecret = oauthClientSecret
		log.Printf("[INFO]: OAuth client secret information loaded from env.")
	}

	adminUsername, ok := os.LookupEnv("admin_username")
	if ok {
		adminUsername = decodeBase64(adminUsername)
		AuthConfig.AdminUsername = adminUsername
		log.Printf("[INFO]: Admin username information loaded from env.")
	}

	adminPassword, ok := os.LookupEnv("admin_password")
	if ok {
		adminPassword = decodeBase64(adminPassword)
		AuthConfig.AdminPassword = adminPassword
		log.Printf("[INFO]: Admin password information loaded from env.")
	}
}

// decodeBase64 Decode base64 string
func decodeBase64(encodedString string) string {

	base64Value, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		fmt.Println("[ERROR] decode secret base64 value with value:  ", encodedString, " - ", err.Error())
		panic(err)
	}
	return string(base64Value)
}
