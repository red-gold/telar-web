package config

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"

	coreUtils "github.com/red-gold/telar-core/utils"
)

const (
	basePath         = "/var/openfaas/secrets/"
	storageSecretKey = "serviceAccountKey.json"
)

var secretKeys = []string{storageSecretKey}

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

	bucketName, ok := os.LookupEnv("bucket_name")
	if ok {
		StorageConfig.BucketName = bucketName
		log.Printf("[INFO]: Public key path information loaded from env.")
	}
}

// loadSecretsFromFile Load secrets from file
func loadSecretsFromFile() {
	filesConfig := getAllConfigFromFile()
	if filesConfig[basePath+storageSecretKey] != nil {
		storageSecret := string(filesConfig[basePath+storageSecretKey])
		StorageConfig.StorageSecret = storageSecret
		log.Printf("[INFO]: OAuth client secret information loaded from env.")
	}

}

// loadSecretsFromEnv Load secrets from environment variables
func loadSecretsFromEnv() {
	storageSecret, ok := os.LookupEnv("service_account_key_json")
	if ok {
		storageSecret = decodeBase64(storageSecret)
		StorageConfig.StorageSecret = storageSecret
		log.Printf("[INFO]: OAuth client secret information loaded from env.")
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
