package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alexellis/hmac"
	"github.com/gofrs/uuid"
	coreConfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/types"
	"github.com/red-gold/telar-core/utils"
	notifyConfig "github.com/red-gold/telar-web/micros/notifications/config"
	"github.com/red-gold/telar-web/micros/notifications/dto"
	"github.com/red-gold/telar-web/micros/notifications/models"
)

const (
	likeNotifyType          = "like"
	followNotifyType        = "follow"
	commentNotifyType       = "comment"
	sendEmailOnLike         = "send_email_on_like"
	sendEmailOnFollow       = "send_email_on_follow"
	sendEmailOnComment      = "send_email_on_comment_post"
	notificationSettingType = "notification"
)

var settingMappedFromNotify = map[string]string{
	likeNotifyType:    sendEmailOnLike,
	followNotifyType:  sendEmailOnFollow,
	commentNotifyType: sendEmailOnComment,
}

type UserInfoInReq struct {
	UserId      uuid.UUID `json:"userId"`
	Username    string    `json:"username"`
	Avatar      string    `json:"avatar"`
	DisplayName string    `json:"displayName"`
	SystemRole  string    `json:"systemRole"`
}

// getSettingPath
func getSettingPath(userId uuid.UUID, settingType, settingKey string) string {
	key := fmt.Sprintf("%s:%s:%s", userId.String(), settingType, settingKey)
	return key
}

// functionCall send request to another function/microservice using HMAC validation
func functionCall(method string, bytesReq []byte, url string, header map[string][]string) ([]byte, error) {
	prettyURL := utils.GetPrettyURLf(url)
	bodyReader := bytes.NewBuffer(bytesReq)

	httpReq, httpErr := http.NewRequest(method, *coreConfig.AppConfig.InternalGateway+prettyURL, bodyReader)
	if httpErr != nil {
		return nil, httpErr
	}

	payloadSecret := *coreConfig.AppConfig.PayloadSecret

	digest := hmac.Sign(bytesReq, []byte(payloadSecret))
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
		return nil, fmt.Errorf("failed to call admin check api, invalid status: %s", res.Status)
	}

	return resData, nil
}

func combineURL(a, b string) string {
	if !strings.HasSuffix(a, "/") {
		a = a + "/"
	}
	if strings.HasPrefix(b, "/") {
		b = strings.TrimPrefix(b, "/")
	}

	return a + b
}

// getUsersSettings Get users settings
func getUsersNotificationSettings(userIds []uuid.UUID, userInfoInReq *UserInfoInReq) (map[string]string, error) {
	url := "/setting/dto/ids"
	model := models.GetSettingsModel{
		UserIds: userIds,
		Type:    "notification",
	}
	payload, marshalErr := json.Marshal(model)
	if marshalErr != nil {
		return nil, marshalErr
	}
	// Create user headers for http request
	userHeaders := make(map[string][]string)
	userHeaders["uid"] = []string{userInfoInReq.UserId.String()}
	userHeaders["email"] = []string{userInfoInReq.Username}
	userHeaders["avatar"] = []string{userInfoInReq.Avatar}
	userHeaders["displayName"] = []string{userInfoInReq.DisplayName}
	userHeaders["role"] = []string{userInfoInReq.SystemRole}

	resData, callErr := functionCall(http.MethodPost, payload, url, userHeaders)
	if callErr != nil {

		return nil, fmt.Errorf("Cannot send request to %s - %s", url, callErr.Error())
	}

	var parsedData map[string]string
	json.Unmarshal(resData, &parsedData)
	return parsedData, nil
}

// sendEmailNotification Send email notification
func sendEmailNotification(model dto.Notification) error {

	email := utils.NewEmail(*coreConfig.AppConfig.RefEmail, *coreConfig.AppConfig.RefEmailPass, *coreConfig.AppConfig.SmtpEmail)
	title := ""
	switch model.Type {
	case likeNotifyType:
		title = fmt.Sprintf("%s liked your post.", model.OwnerDisplayName)
	case commentNotifyType:
		title = fmt.Sprintf("%s  added a comment on your post.", model.OwnerDisplayName)
	case followNotifyType:
		title = fmt.Sprintf("%s  now following you.", model.OwnerDisplayName)
	}

	subject := fmt.Sprintf("%s Notification - %s", *coreConfig.AppConfig.AppName, title)
	emailReq := utils.NewEmailRequest([]string{model.NotifyRecieverEmail}, subject, "")

	emailResStatus, emailResErr := email.SendEmail(emailReq, "views/notify_email.html", struct {
		AppName         string
		AppURL          string
		Title           string
		Avatar          string
		FullName        string
		ViewLink        string
		UnsubscribeLink string
	}{
		AppName:         *coreConfig.AppConfig.AppName,
		AppURL:          notifyConfig.NotificationConfig.WebURL,
		Title:           title,
		Avatar:          model.OwnerAvatar,
		FullName:        model.OwnerDisplayName,
		ViewLink:        combineURL(notifyConfig.NotificationConfig.WebURL, model.URL),
		UnsubscribeLink: combineURL(notifyConfig.NotificationConfig.WebURL, "settings/notify"),
	})

	if emailResErr != nil {
		return fmt.Errorf("Error happened in sending email error: %s", emailResErr.Error())

	}
	if !emailResStatus {
		return fmt.Errorf("Email response status is false! ")

	}
	return nil
}
