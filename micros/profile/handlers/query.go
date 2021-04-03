package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	service "github.com/red-gold/telar-web/micros/profile/services"
)

// QueryUserProfileHandle handle queru on userProfile
func QueryUserProfileHandle(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		// Create service
		userService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %s", err)

		}

		searchParam := r.Form.Get("search")
		pageParam := r.Form.Get("page")
		page := 0
		if pageParam != "" {
			var strErr error
			page, strErr = strconv.Atoi(pageParam)
			if strErr != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, strErr
			}
		}
		var nin []uuid.UUID
		for _, v := range r.Form["nin"] {
			parsedUUID, uuidErr := uuid.FromString(v)

			if uuidErr != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, uuidErr
			}

			nin = append(nin, parsedUUID)
		}
		userList, err := userService.QueryUserProfile(searchParam, "created_date", int64(page), nin)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		body, marshalErr := json.Marshal(userList)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("Error while marshaling userList: %s",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("userListMarshalError", errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}
