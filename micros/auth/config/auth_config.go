package config

import "time"

type (
	Configuration struct {
		OAuthProvider          string
		OAuthProviderBaseURL   string
		OAuthTelarBaseURL      string
		ClientID               string
		ClientSecret           string
		OAuthClientSecret      string
		AdminUsername          string
		AdminPassword          string
		ExternalRedirectDomain string
		AuthWebURI             string
		WebURL                 string
		Scope                  string
		CookieRootDomain       string
		CookieExpiresIn        time.Duration
		BaseRoute              string
		VerifyType             string
		QueryPrettyURL         bool
		Debug                  bool // Debug enables verbose logging of claims / cookies
	}
)

// AuthConfig holds the configuration values from auth-config.yml file
var AuthConfig Configuration
