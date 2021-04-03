package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	firebase "firebase.google.com/go"
	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	appConfig "github.com/red-gold/telar-web/micros/storage/config"
	"google.golang.org/api/option"

	coreSetting "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
)

// UploadeHandle a function invocation
func UploadeHandle() func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {
	ctx := context.Background()

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {

		storageConfig := &appConfig.StorageConfig
		fmt.Printf("\n Upload userId : %v\n", req.UserID)

		fmt.Println("File Upload Endpoint Hit")
		// params from /storage/:uid/:dir
		dirName := req.GetParamByName("dir")
		if dirName == "" {
			errorMessage := fmt.Sprintf("Directory name is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("dirNameRequired", errorMessage)}, nil
		}
		fmt.Printf("\n Directory name: %s", dirName)

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		r.ParseMultipartForm(10 << 20)
		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		file, handlerFile, err := r.FormFile("file")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)

		}
		defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", handlerFile.Filename)
		fmt.Printf("File Size: %+v\n", handlerFile.Size)
		fmt.Printf("MIME Header: %+v\n", handlerFile.Header)

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
			log.Fatalln(err)
		}

		client, err := app.Storage(ctx)
		if err != nil {
			log.Fatalln(err)
		}

		bucket, err := client.DefaultBucket()
		if err != nil {
			log.Fatalln(err)
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
