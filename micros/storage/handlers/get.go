package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/utils"
	appConfig "github.com/red-gold/telar-web/micros/storage/config"
)

var redisClient *redis.Client

func init() {

}

// GetFileHandle a function invocation
func GetFileHandle(c *fiber.Ctx) error {

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

	log.Info("File Upload Endpoint Hit")

	dirName := c.Params("dir")
	if dirName == "" {
		errorMessage := fmt.Sprintf("Directory name is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("dirNameRequired", "Directory name is required!"))
	}

	log.Info("Directory name: %s", dirName)

	fileName := c.Params("name")
	if fileName == "" {
		errorMessage := fmt.Sprintf("File name is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("fileNameRequired", "File name is required!"))
	}

	log.Info("File name: %s", fileName)

	userId := c.Params("uid")
	if userId == "" {
		errorMessage := fmt.Sprintf("User Id is required!")
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("fileNameRequired", "User id is required!"))
	}

	log.Info("\n User ID: %s", userId)
	userUUID, uuidErr := uuid.FromString(userId)
	if uuidErr != nil {
		errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
		log.Error(errorMessage)
		return c.Status(http.StatusBadRequest).JSON(utils.Error("uuidError", "can not parseUser id!"))
	}

	objectName := fmt.Sprintf("%s/%s/%s", userUUID, dirName, fileName)

	// Generate download URL
	downloadURL, urlErr := generateV4GetObjectSignedURL(storageConfig.BucketName, objectName, storageConfig.StorageSecret)
	if urlErr != nil {
		fmt.Println(urlErr.Error())
	}

	cacheSince := time.Now().Format(http.TimeFormat)
	cacheUntil := time.Now().Add(time.Second * time.Duration(cacheTimeout)).Format(http.TimeFormat)

	c.Set("Cache-Control", fmt.Sprintf("max-age:%d, public", cacheTimeout))
	c.Set("Last-Modified", cacheSince)
	c.Set("Expires", cacheUntil)
	return c.Redirect(downloadURL, http.StatusTemporaryRedirect)

}
