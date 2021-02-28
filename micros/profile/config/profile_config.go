package config

type (
	Configuration struct {
		BaseRoute      string
		QueryPrettyURL bool
		Debug          bool // Debug enables verbose logging of claims / cookies
	}
)

// ProfileConfig holds the configuration values from profile-config.yml file
var ProfileConfig Configuration
