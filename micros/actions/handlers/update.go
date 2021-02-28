package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	domain "github.com/red-gold/telar-web/micros/actions/dto"
	models "github.com/red-gold/telar-web/micros/actions/models"
	service "github.com/red-gold/telar-web/micros/actions/services"
)

// UpdateActionRoomHandle handle create a new actionRoom
func UpdateActionRoomHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create the model object
		var model models.ActionRoomModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, err
		}

		// Create service
		actionRoomService, serviceErr := service.NewActionRoomService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		updatedActionRoom := &domain.ActionRoom{
			ObjectId:    model.ObjectId,
			OwnerUserId: req.UserID,
			PrivateKey:  model.PrivateKey,
			AccessKey:   model.AccessKey,
			Status:      model.Status,
			CreatedDate: model.CreatedDate,
		}

		if err := actionRoomService.UpdateActionRoomById(updatedActionRoom); err != nil {
			errorMessage := fmt.Sprintf("Update ActionRoom Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("updateActionRoomError", errorMessage)}, nil
		}
		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}

// SetAccessKeyHandle handle create a new actionRoom
func SetAccessKeyHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create service
		actionRoomService, serviceErr := service.NewActionRoomService(db)
		if serviceErr != nil {
			return handler.Response{StatusCode: http.StatusInternalServerError}, serviceErr
		}

		accessKey, err := actionRoomService.SetAccessKey(req.UserID)
		if err != nil {
			errorMessage := fmt.Sprintf("Update ActionRoom Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("updateActionRoomError", errorMessage)}, nil
		}
		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true, "accessKey": "%s"}`, accessKey)),
			StatusCode: http.StatusOK,
		}, nil
	}
}
