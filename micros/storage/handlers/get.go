package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	"github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	appConfig "github.com/red-gold/telar-web/micros/storage/config"
)

var redisClient *redis.Client

func init() {

}

// GetFileHandle a function invocation
func GetFileHandle(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {

		storageConfig := &appConfig.StorageConfig

		// Initialize Redis Connection
		if redisClient == nil {

			redisPassword, redisErr := utils.ReadSecret("redis-pwd")

			if redisErr != nil {
				fmt.Printf("\n\ncouldn't get payload-secret: %s\n\n", redisErr.Error())
			}
			fmt.Println(storageConfig.RedisAddress)
			fmt.Println(redisPassword)
			redisClient = redis.NewClient(&redis.Options{
				Addr:     storageConfig.RedisAddress,
				Password: redisPassword,
				DB:       0,
			})
			pong, err := redisClient.Ping().Result()
			fmt.Println(pong, err)
		}

		fmt.Println("File Upload Endpoint Hit")
		// params from /storage/:uid/:dir/:name
		dirName := req.GetParamByName("dir")
		if dirName == "" {
			errorMessage := fmt.Sprintf("Directory name is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("dirNameRequired", errorMessage)}, nil
		}
		fmt.Printf("\n Directory name: %s", dirName)

		// params from /storage/:dir
		fileName := req.GetParamByName("name")
		if fileName == "" {
			errorMessage := fmt.Sprintf("File name is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("fileNameRequired", errorMessage)}, nil
		}
		fmt.Printf("\n File name: %s", fileName)

		userId := req.GetParamByName("uid")
		if userId == "" {
			errorMessage := fmt.Sprintf("User Id is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userIdRequired", errorMessage)}, nil
		}

		fmt.Printf("\n User ID: %s", userId)
		userUUID, uuidErr := uuid.FromString(userId)
		if uuidErr != nil {
			errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("uuidError", errorMessage)}, nil
		}

		objectName := fmt.Sprintf("%s/%s/%s", userUUID, dirName, fileName)

		// Generate download URL
		downloadURL, urlErr := generateV4GetObjectSignedURL(storageConfig.BucketName, objectName, storageConfig.StorageSecretPath)
		if urlErr != nil {
			fmt.Println(urlErr.Error())
		}

		code := 302 // Permanent redirect, request with GET method
		if r.Method != http.MethodGet {
			// Temporary redirect, request with same method
			// As of Go 1.3, Go does not support status code 308.
			code = 307
		}
		cacheSince := time.Now().Format(http.TimeFormat)
		cacheUntil := time.Now().Add(time.Second * time.Duration(cacheTimeout)).Format(http.TimeFormat)
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age:%d, public", cacheTimeout))
		w.Header().Set("Last-Modified", cacheSince)
		w.Header().Set("Expires", cacheUntil)
		http.Redirect(w, r, downloadURL, code)

		return handler.Response{
			StatusCode: http.StatusTemporaryRedirect,
		}, nil

	}

}
