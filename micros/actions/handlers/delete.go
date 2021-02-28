package handlers

import (
	"fmt"
	"net/http"

	uuid "github.com/gofrs/uuid"
	handler "github.com/openfaas-incubator/go-function-sdk"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	service "github.com/red-gold/telar-web/micros/actions/services"
)

// DeleteActionRoomHandle handle delete a ActionRoom
func DeleteActionRoomHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {

		// params from /actions/room/:roomId
		actionRoomId := req.GetParamByName("roomId")
		if actionRoomId == "" {
			errorMessage := fmt.Sprintf("ActionRoom Id is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("actionRoomIdRequired", errorMessage)}, nil
		}
		fmt.Printf("\n ActionRoom ID: %s", actionRoomId)
		actionRoomUUID, uuidErr := uuid.FromString(actionRoomId)
		if uuidErr != nil {
			errorMessage := fmt.Sprintf("UUID Error %s", uuidErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("uuidError", errorMessage)}, nil
		}
		fmt.Printf("\n ActionRoom UUID: %s", actionRoomUUID)
		// Create service
		actionRoomService, serviceErr := service.NewActionRoomService(db)
		if serviceErr != nil {
			errorMessage := fmt.Sprintf("ActionRoom Service Error %s", serviceErr.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("actionRoomServiceError", errorMessage)}, nil

		}

		if err := actionRoomService.DeleteActionRoomByOwner(req.UserID, actionRoomUUID); err != nil {
			errorMessage := fmt.Sprintf("Delete ActionRoom Error %s", err.Error())
			return handler.Response{StatusCode: http.StatusInternalServerError, Body: utils.MarshalError("deleteActionRoomError", errorMessage)}, nil

		}
		return handler.Response{
			Body:       []byte(`{"success": true}`),
			StatusCode: http.StatusOK,
		}, nil
	}
}
