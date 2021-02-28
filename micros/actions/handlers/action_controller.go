package handlers

import (
	"bytes"
	"fmt"
	"net/http"

	handler "github.com/openfaas-incubator/go-function-sdk"
	coreServer "github.com/red-gold/telar-core/server"
	server "github.com/red-gold/telar-core/server"
	"github.com/red-gold/telar-core/utils"
	serviceConfig "github.com/red-gold/telar-web/micros/actions/config"
)

// DispatchHandle handle create a new actionRoom
func DispatchHandle(db interface{}) func(server.Request) (handler.Response, error) {

	return func(req server.Request) (handler.Response, error) {
		actionConfig := serviceConfig.ActionConfig

		// params from /actions/dispatch/:roomId
		actionRoomId := req.GetParamByName("roomId")
		if actionRoomId == "" {
			errorMessage := fmt.Sprintf("ActionRoom Id is required!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("actionRoomIdRequired", errorMessage)}, nil
		}

		bodyReader := bytes.NewBuffer(req.Body)
		URL := fmt.Sprintf("%s/api/dispatch/%s", actionConfig.WebsocketServerURL, actionRoomId)
		fmt.Printf("\n\n Dispatch URL: %s\n\n", URL)
		httpReq, httpErr := http.NewRequest(http.MethodPost, URL, bodyReader)
		if httpErr != nil {
			errorMessage := fmt.Sprintf("Error while creating dispatch request!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("createDispatchRequestError", errorMessage)}, nil
		}

		xCloudSignature := req.Header.Get(coreServer.X_Cloud_Signature)
		httpReq.Header.Add(coreServer.X_Cloud_Signature, xCloudSignature)

		c := http.Client{}
		res, reqErr := c.Do(httpReq)
		if res.Body != nil {
			defer res.Body.Close()
		}
		if reqErr != nil {
			errorMessage := fmt.Sprintf("Error while sending dispatch request to websocket server!")
			return handler.Response{StatusCode: http.StatusBadRequest, Body: utils.MarshalError("sendDispatchRequestError", errorMessage)}, nil
		}

		return handler.Response{
			Body:       []byte(fmt.Sprintf(`{"success": true}`)),
			StatusCode: http.StatusOK,
		}, nil
	}
}
