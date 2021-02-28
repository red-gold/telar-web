package config

import (
	"log"
	"os"
	"strconv"
)

// Initialize AppConfig
func InitConfig() {

	// Load config from environment values if exists
	loadConfigFromEnvironment()
}

// Load config from environment
func loadConfigFromEnvironment() {

	base_route, ok := os.LookupEnv("base_route")
	if ok {
		StorageConfig.BaseRoute = base_route
		log.Printf("[INFO]: Base route information loaded from env.")
	}

	queryPrettyURL, ok := os.LookupEnv("query_pretty_url")
	if ok {
		parsedQueryPrettyURL, errParseDebug := strconv.ParseBool(queryPrettyURL)
		if errParseDebug != nil {
			log.Printf("[ERROR]: Query Pretty URL information loading error: %s", errParseDebug.Error())
		}
		StorageConfig.QueryPrettyURL = parsedQueryPrettyURL
		log.Printf("[INFO]: Query Pretty URL information loaded from env.")
	}

	redisAddres, ok := os.LookupEnv("redis_address")
	if ok {

		StorageConfig.RedisAddress = redisAddres
		log.Printf("[INFO]: Redis Address URL information loaded from env: %s", redisAddres)
	}

	externalDomain, ok := os.LookupEnv("external_domain")
	if ok {
		StorageConfig.ExternalDomain = externalDomain
		log.Printf("[INFO]: External domain information loaded from env.")
	}

	debug, ok := os.LookupEnv("write_debug")
	if ok {
		parsedDebug, errParseDebug := strconv.ParseBool(debug)
		if errParseDebug != nil {
			log.Printf("[ERROR]: Debug information loading error: %s", errParseDebug.Error())
		}
		StorageConfig.Debug = parsedDebug
		log.Printf("[INFO]: Debug information loaded from env.")
	}

	storageSecretPath, ok := os.LookupEnv("storage_secret_path")
	if ok {
		StorageConfig.StorageSecretPath = storageSecretPath
		log.Printf("[INFO]: Public key path information loaded from env.")
	}

	bucketName, ok := os.LookupEnv("bucket_name")
	if ok {
		StorageConfig.BucketName = bucketName
		log.Printf("[INFO]: Public key path information loaded from env.")
	}
}
