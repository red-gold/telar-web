package handlers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alexellis/hmac"
	coreConfig "github.com/red-gold/telar-core/config"
	ac "github.com/red-gold/telar-web/micros/admin/config"

	server "github.com/red-gold/telar-core/server"
	utils "github.com/red-gold/telar-core/utils"
)

// functionCall send request to another function/microservice using HMAC validation
func functionCall(bytesReq []byte, url string) ([]byte, error) {
	prettyURL := utils.GetPrettyURLf(url)
	bodyReader := bytes.NewBuffer(bytesReq)

	httpReq, httpErr := http.NewRequest(http.MethodPost, *coreConfig.AppConfig.InternalGateway+prettyURL, bodyReader)
	if httpErr != nil {
		return nil, httpErr
	}

	payloadSecret := *coreConfig.AppConfig.PayloadSecret

	digest := hmac.Sign(bytesReq, []byte(payloadSecret))
	httpReq.Header.Set("Content-type", "application/json")
	fmt.Printf("\ndigest: %s, header: %v \n", "sha1="+hex.EncodeToString(digest), server.X_Cloud_Signature)
	httpReq.Header.Add(server.X_Cloud_Signature, "sha1="+hex.EncodeToString(digest))

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
		return nil, fmt.Errorf("failed to call %s api, invalid status: %s", prettyURL, res.Status)
	}

	return resData, nil
}

// functionCallByCookie send request to another function/microservice using cookie validation
func functionCallByHeader(method string, bytesReq []byte, url string, header map[string][]string) ([]byte, error) {
	prettyURL := utils.GetPrettyURLf(url)
	bodyReader := bytes.NewBuffer(bytesReq)

	httpReq, httpErr := http.NewRequest(method, *coreConfig.AppConfig.InternalGateway+prettyURL, bodyReader)
	if httpErr != nil {
		return nil, httpErr
	}
	payloadSecret := *coreConfig.AppConfig.PayloadSecret

	digest := hmac.Sign(bytesReq, []byte(payloadSecret))
	httpReq.Header.Set("Content-type", "application/json")
	fmt.Printf("\ndigest: %s, header: %v \n", "sha1="+hex.EncodeToString(digest), server.X_Cloud_Signature)
	httpReq.Header.Add(server.X_Cloud_Signature, "sha1="+hex.EncodeToString(digest))
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
		return nil, fmt.Errorf("failed to call %s api, invalid status: %s", prettyURL, res.Status)
	}

	return resData, nil
}

// writeTokenOnCookie wite session on cookie
func writeSessionOnCookie(w http.ResponseWriter, session string, adminConfig *ac.Configuration) {
	appConfig := coreConfig.AppConfig
	parts := strings.Split(session, ".")
	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Name:     *appConfig.HeaderCookieName,
		Value:    parts[0],
		Path:     "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: adminConfig.CookieRootDomain,
	})

	http.SetCookie(w, &http.Cookie{
		// HttpOnly: true,
		Name:  *appConfig.PayloadCookieName,
		Value: parts[1],
		Path:  "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: adminConfig.CookieRootDomain,
	})

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Name:     *appConfig.SignatureCookieName,
		Value:    parts[2],
		Path:     "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: adminConfig.CookieRootDomain,
	})
}
