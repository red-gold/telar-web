package config

type (
	Configuration struct {
		BaseRoute        string
		CookieRootDomain string
		QueryPrettyURL   bool
		Debug            bool // Debug enables verbose logging of claims / cookies
	}
)

// AdminConfig holds the configuration values from admin-config.yml file
var AdminConfig Configuration
