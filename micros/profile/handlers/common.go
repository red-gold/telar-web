package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/alexellis/hmac"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	coreConfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/pkg/parser"
	"github.com/red-gold/telar-core/types"
	utils "github.com/red-gold/telar-core/utils"
	models "github.com/red-gold/telar-web/micros/profile/models"
)

type Action struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type UserInfoInReq struct {
	UserId      uuid.UUID `json:"userId"`
	Username    string    `json:"username"`
	Avatar      string    `json:"avatar"`
	DisplayName string    `json:"displayName"`
	SystemRole  string    `json:"systemRole"`
}

// contains
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

type UpdateProfileQueryModel struct {
	UpdateType int `query:"updateType"`
}

const UpdateAllType = 0
const UpdateGeneralType = 1
const UpdateSocialInfoType = 2

// getHeadersFromUserInfoReq
func getHeadersFromUserInfoReq(info *UserInfoInReq) map[string][]string {
	userHeaders := make(map[string][]string)
	userHeaders["uid"] = []string{info.UserId.String()}
	userHeaders["email"] = []string{info.Username}
	userHeaders["avatar"] = []string{info.Avatar}
	userHeaders["displayName"] = []string{info.DisplayName}
	userHeaders["role"] = []string{info.SystemRole}

	return userHeaders
}

// getUserInfoReq
func getUserInfoReq(c *fiber.Ctx) *UserInfoInReq {
	currentUser, ok := c.Locals("user").(types.UserContext)
	if !ok {
		return &UserInfoInReq{}
	}
	userInfoInReq := &UserInfoInReq{
		UserId:      currentUser.UserID,
		Username:    currentUser.Username,
		Avatar:      currentUser.Avatar,
		DisplayName: currentUser.DisplayName,
		SystemRole:  currentUser.SystemRole,
	}
	return userInfoInReq

}

// Dispatch action
func dispatchAction(action Action, userInfoInReq *UserInfoInReq) {

	actionURL := fmt.Sprintf("/actions/dispatch/%s", userInfoInReq.UserId.String())

	actionBytes, marshalErr := json.Marshal(action)
	if marshalErr != nil {
		errorMessage := fmt.Sprintf("Marshal notification Error %s", marshalErr.Error())
		fmt.Println(errorMessage)
	}
	// Create user headers for http request
	userHeaders := make(map[string][]string)
	userHeaders["uid"] = []string{userInfoInReq.UserId.String()}
	userHeaders["email"] = []string{userInfoInReq.Username}
	userHeaders["avatar"] = []string{userInfoInReq.Avatar}
	userHeaders["displayName"] = []string{userInfoInReq.DisplayName}
	userHeaders["role"] = []string{userInfoInReq.SystemRole}

	_, actionErr := functionCall(http.MethodPost, actionBytes, actionURL, userHeaders)

	if actionErr != nil {
		errorMessage := fmt.Sprintf("Cannot send action request! error: %s", actionErr.Error())
		fmt.Println(errorMessage)
	}
}

// functionCall send request to another function/microservice using HMAC validation
func functionCall(method string, bytesReq []byte, url string, header map[string][]string) ([]byte, error) {
	prettyURL := utils.GetPrettyURLf(url)
	bodyReader := bytes.NewBuffer(bytesReq)

	httpReq, httpErr := http.NewRequest(method, *coreConfig.AppConfig.InternalGateway+prettyURL, bodyReader)
	if httpErr != nil {
		return nil, httpErr
	}

	digest := hmac.Sign(bytesReq, []byte(*coreConfig.AppConfig.PayloadSecret))
	httpReq.Header.Set("Content-type", "application/json")
	fmt.Printf("\ndigest: %s, header: %v \n", "sha1="+hex.EncodeToString(digest), types.HeaderHMACAuthenticate)
	httpReq.Header.Add(types.HeaderHMACAuthenticate, "sha1="+hex.EncodeToString(digest))

	if header != nil {
		for k, v := range header {
			httpReq.Header[k] = v
		}
	}

	c := http.Client{}
	res, reqErr := c.Do(httpReq)
	fmt.Printf("\nRes: %v\n", res)
	if reqErr != nil {
		return nil, fmt.Errorf("Error while sending admin check request!: %s", reqErr.Error())
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	resData, readErr := ioutil.ReadAll(res.Body)
	if resData == nil || readErr != nil {
		return nil, fmt.Errorf("failed to read response from admin check request.")
	}

	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, NotFoundHTTPStatusError
		}
		return nil, fmt.Errorf("failed to call %s api, invalid status: %s", prettyURL, res.Status)
	}

	return resData, nil
}

// getUpdateModel
func getUpdateModel(c *fiber.Ctx) (interface{}, error) {

	query := new(UpdateProfileQueryModel)

	if err := parser.QueryParser(c, query); err != nil {
		return nil, err
	}

	if query.UpdateType == UpdateSocialInfoType {
		model := new(models.SocialInfoUpdateModel)
		unmarshalErr := c.BodyParser(model)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}

		model.LastUpdated = utils.UTCNowUnix()

		return model, nil
	} else if query.UpdateType == UpdateGeneralType {
		model := new(models.ProfileGeneralUpdateModel)
		unmarshalErr := c.BodyParser(model)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		model.LastUpdated = utils.UTCNowUnix()
		return model, nil
	}
	model := new(models.ProfileUpdateModel)
	unmarshalErr := c.BodyParser(model)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	model.LastUpdated = utils.UTCNowUnix()
	return model, nil

}
