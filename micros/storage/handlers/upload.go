package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	firebase "firebase.google.com/go"
	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	appConfig "github.com/red-gold/telar-web/micros/storage/config"
	"google.golang.org/api/option"

	coreSetting "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/log"
	"github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
)

// UploadeHandle a function invocation
func UploadeHandle() func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	ctx := context.Background()

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {

		storageConfig := &appConfig.StorageConfig
		log.Info("Hit upload endpoint by userId : %v", req.UserID)

		// params from /storage/:uid/:dir
		dirName := req.GetParamByName("dir")
		if dirName == "" {
			errorMessage := fmt.Sprintf("Directory name is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("dirNameRequired", errorMessage)}, nil
		}
		log.Info("Directory name: %s", dirName)

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		log.Info("Start parsing multipartform")
		parseError := r.ParseMultipartForm(1024 * 10)
		if parseError != nil {
			log.Error("Parsing multipartform %s", parseError.Error())
			return handler.Response{
				Body:       utils.MarshalError("upload/internal", "Parsing multipartform"),
				StatusCode: http.StatusNotAcceptable,
			}, nil
		}

		// FormFile returns the first file for the given key `file`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		file, handlerFile, err := r.FormFile("file")
		if err != nil {
			log.Error("Error Retrieving the File %s", err.Error())
			return handler.Response{
				Body:       utils.MarshalError("upload/internal", "Error Retrieving the File"),
				StatusCode: http.StatusInternalServerError,
			}, nil

		}
		log.Info("Parsed file %v", file)
		log.Info("Uploaded File: %+v", handlerFile.Filename)
		log.Info("File Size: %+v", handlerFile.Size)
		log.Info("MIME Header: %+v", handlerFile.Header)

		defer file.Close()
		extension := filepath.Ext(handlerFile.Filename)
		fileNameUUID, uuidErr := uuid.NewV4()
		if uuidErr != nil {
			errorMessage := fmt.Sprintf("File name from UUID error: %s", uuidErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("fileNameUUIDError", errorMessage)},
				nil
		}

		fileName := fileNameUUID.String()
		fileNameWithExtension := fmt.Sprintf("%s%s", fileName, extension)

		objectName := fmt.Sprintf("%s/%s/%s", req.UserID.String(), dirName, fileNameWithExtension)
		config := &firebase.Config{
			StorageBucket: storageConfig.BucketName,
		}

		opt := option.WithCredentialsJSON([]byte(storageConfig.StorageSecret))
		app, err := firebase.NewApp(ctx, config, opt)
		if err != nil {
			log.Error("Credential parse %s", err.Error())
			return handler.Response{
				Body:       utils.MarshalError("upload/internal", "Credential parse"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		client, err := app.Storage(ctx)
		if err != nil {
			log.Error("Get storage client %s", err.Error())
			return handler.Response{
				Body:       utils.MarshalError("upload/internal", "Get storage client"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		bucket, err := client.DefaultBucket()
		if err != nil {
			log.Error("Get default bucket %s", err.Error())
			return handler.Response{
				Body:       utils.MarshalError("upload/internal", "Get default bucket"),
				StatusCode: http.StatusInternalServerError,
			}, nil

		}

		wc := bucket.Object(objectName).NewWriter(ctx)
		if _, err = io.Copy(wc, file); err != nil {
			fmt.Println(err.Error())
		}
		if err := wc.Close(); err != nil {
			fmt.Println(err.Error())
		}

		prettyURL := utils.GetPrettyURLf(storageConfig.BaseRoute)
		downloadURL := fmt.Sprintf("%s/%s/%s/%s", *coreSetting.AppConfig.Gateway+prettyURL, req.UserID, dirName, fileNameWithExtension)
		return handler.Response{
			Body:       []byte(fmt.Sprintf("{ \"success\": true, \"payload\": { \"url\": \"%s\"}}", downloadURL)),
			StatusCode: http.StatusOK,
		}, nil
	}

}
