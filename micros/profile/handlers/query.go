package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	service "github.com/red-gold/telar-web/micros/auth/services"
)

// QueryUserProfileHandle handle queru on userProfile
func QueryUserProfileHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		userService, serviceErr := service.NewUserProfileService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		var query *url.Values
		if len(req.QueryString) > 0 {
			q, err := url.ParseQuery(string(req.QueryString))
			if err != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, err
			}
			query = &q
		}
		searchParam := query.Get("search")
		pageParam := query.Get("page")
		page := 0
		if pageParam != "" {
			var strErr error
			page, strErr = strconv.Atoi(pageParam)
			if strErr != nil {
				return handler.Response{StatusCode: http.StatusInternalServerError}, strErr
			}
		}
		userList, err := userService.QueryUserProfile(searchParam, "created_date", int64(page))
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
