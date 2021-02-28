package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
	models "github.com/red-gold/telar-web/micros/actions/models"
	service "github.com/red-gold/telar-web/micros/actions/services"
)

// GetActionRoomHandle handle get a actionRoom
func GetActionRoomHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		actionRoomId := req.GetParamByName("actionRoomId")
		actionRoomUUID, uuidErr := uuid.FromString(actionRoomId)
		if uuidErr != nil {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("parseUUIDError", "Can not parse actionRoom id!")},
				nil
		}
		// Create service
		actionRoomService, serviceErr := service.NewActionRoomService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		foundActionRoom, err := actionRoomService.FindById(actionRoomUUID)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		actionRoomModel := models.ActionRoomModel{
			ObjectId:    foundActionRoom.ObjectId,
			OwnerUserId: foundActionRoom.OwnerUserId,
			PrivateKey:  foundActionRoom.PrivateKey,
			AccessKey:   foundActionRoom.AccessKey,
			Status:      foundActionRoom.Status,
			CreatedDate: foundActionRoom.CreatedDate,
		}

		body, marshalErr := json.Marshal(actionRoomModel)
		if marshalErr != nil {
			errorMessage := fmt.Sprintf("{error: 'Error while marshaling actionRoomModel: %s'}",
				marshalErr.Error())
			return handler.Response{StatusCode: http.StatusBadRequest, Body: []byte(errorMessage)},
				marshalErr

		}
		return handler.Response{
			Body:       []byte(body),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// GetAccessKeyHandle handle get access key
func GetAccessKeyHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		// Create service
		actionRoomService, serviceErr := service.NewActionRoomService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		accessKey, err := actionRoomService.GetAccessKey(req.UserID)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf("{\"success\": true, \"accessKey\": \"%s\"}", accessKey)),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// VerifyAccessKeyHandle handle verify access key
func VerifyAccessKeyHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// actions/room/verify/:accessKey
		accessKey := req.GetParamByName("accessKey")
		if accessKey == "" {
			return handler.Response{StatusCode: http.StatusBadRequest,
					Body: utils.MarshalError("accessKeyRequiredError", "Access key is required!")},
				nil
		}

		// Create service
		actionRoomService, serviceErr := service.NewActionRoomService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		isVerified, err := actionRoomService.VerifyAccessKey(req.UserID, accessKey)
		if err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf("{\"success\": true, \"isVerified\": %t}", isVerified)),
			StatusCode: http.StatusOK,
		}, nil
	}
}
