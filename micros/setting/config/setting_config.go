package config

type (
	Configuration struct {
		BaseRoute      string
		QueryPrettyURL bool
		Debug          bool // Debug enables verbose logging of claims / cookies
	}
)

// UserSettingConfig holds the configuration values from setting-config.yml file
var UserSettingConfig Configuration
