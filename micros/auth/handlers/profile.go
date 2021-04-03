package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	cf "github.com/red-gold/telar-web/micros/auth/config"
	models "github.com/red-gold/telar-web/micros/auth/models"
	"github.com/red-gold/telar-web/micros/auth/provider"
)

// UpdateProfileHandle a function invocation
func UpdateProfileHandle(db interface{}) func(http.ResponseWriter, *http.Request, server.Request) (handler.Response, error) {

	return func(w http.ResponseWriter, r *http.Request, req server.Request) (handler.Response, error) {
		authConfig := &cf.AuthConfig

		model := models.ProfileUpdateModel{}
		unmarshalErr := json.Unmarshal(req.Body, &model)
		if unmarshalErr != nil {
			errorMessage := fmt.Sprintf("Error while un-marshaling ProfileUpdateModel: %s",
				unmarshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("marshalProfileUpdateModelError", errorMessage)},
				unmarshalErr

		}

		err := updateUserProfile(&model, req.UserID, req.Username, req.Avatar, req.DisplayName, req.SystemRole)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}
		tokenModel := &TokenModel{
			token:            ProviderAccessToken{},
			oauthProvider:    nil,
			providerName:     "telar",
			profile:          &provider.Profile{Name: model.FullName, ID: req.UserID.String(), Login: req.Username},
			organizationList: "Red Gold",
			claim: UserClaim{
				DisplayName: model.FullName,
				Email:       req.Username,
				Avatar:      model.Avatar,
				UserId:      req.UserID.String()},
		}
		session, err := createToken(tokenModel)
		if err != nil {
			fmt.Printf("{error: 'Error creating session: %s'}", err.Error())
			return handler.Response{
				Body:       []byte("{error: 'Internal server error creating JWT'}"),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		// Write session on cookie
		writeSessionOnCookie(w, session, authConfig)
		fmt.Printf("\nSession is created: %s \n", session)
		return handler.Response{
			Body:       []byte("{status: true}"),
			StatusCode: http.StatusOK,
		}, nil
	}

}
