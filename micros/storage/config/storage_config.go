package config

type (
	Configuration struct {
		BaseRoute      string
		StorageSecret  string
		ExternalDomain string
		BucketName     string
		RedisAddress   string
		QueryPrettyURL bool
		Debug          bool // Debug enables verbose logging of claims / cookies
	}
)

// StorageConfig holds the configuration values from storage-config.yml file
var StorageConfig Configuration
