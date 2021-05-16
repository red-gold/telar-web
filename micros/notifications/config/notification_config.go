package config

type (
	Configuration struct {
		BaseRoute      string
		WebURL         string
		QueryPrettyURL bool
		Debug          bool // Debug enables verbose logging of claims / cookies
	}
)

// NotificationConfig holds the configuration values from notification-config.yml file
var NotificationConfig Configuration
