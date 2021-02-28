package config

type (
	Configuration struct {
		BaseRoute          string
		WebsocketServerURL string
		QueryPrettyURL     bool
		Debug              bool // Debug enables verbose logging of claims / cookies
	}
)

// ActionConfig holds the configuration values from action-config.yml file
var ActionConfig Configuration
