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

// CreateActionRoomHandle handle create a new actionRoom
func CreateActionRoomHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// Create the model object
		var model models.CreateActionRoomModel
		if err := json.Unmarshal(req.Body, &model); err != nil {
			errorMessage := fmt.Sprintf("Unmarshal CreateActionRoomModel Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("modelMarshalError", errorMessage)}, nil
		}

		// Create service
		actionRoomService, serviceErr := service.NewActionRoomService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("actionRoom Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("actionRoomServiceError", errorMessage)}, nil
		}

		newActionRoom := &domain.ActionRoom{
			ObjectId:    model.ObjectId,
			OwnerUserId: req.UserID,
			PrivateKey:  model.PrivateKey,
			AccessKey:   model.AccessKey,
			Status:      model.Status,
			CreatedDate: model.CreatedDate,
		}

		if err := actionRoomService.SaveActionRoom(newActionRoom); err != nil {
			errorMessage := fmt.Sprintf("Save ActionRoom Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("saveActionRoomError", errorMessage)}, nil
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true, "objectId": "%s"}`, newActionRoom.ObjectId.String())),
			StatusCode: http.StatusOK,
		}, nil
	}
}
