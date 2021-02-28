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
		AdminConfig.BaseRoute = base_route
		log.Printf("[INFO]: Base route information loaded from env.")
	}

	cookieRootDomain, ok := os.LookupEnv("cookie_root_domain")
	if ok {
		AdminConfig.CookieRootDomain = cookieRootDomain
		log.Printf("[INFO]: Cookie root domain information loaded from env.")
	}

	queryPrettyURL, ok := os.LookupEnv("query_pretty_url")
	if ok {
		parsedQueryPrettyURL, errParseDebug := strconv.ParseBool(queryPrettyURL)
		if errParseDebug != nil {
			log.Printf("[ERROR]: Query Pretty URL information loading error: %s", errParseDebug.Error())
		}
		AdminConfig.QueryPrettyURL = parsedQueryPrettyURL
		log.Printf("[INFO]: Query Pretty URL information loaded from env.")
	}
	debug, ok := os.LookupEnv("write_debug")
	if ok {
		parsedDebug, errParseDebug := strconv.ParseBool(debug)
		if errParseDebug != nil {
			log.Printf("[ERROR]: Debug information loading error: %s", errParseDebug.Error())
		}
		AdminConfig.Debug = parsedDebug
		log.Printf("[INFO]: Debug information loaded from env.")
	}
}
