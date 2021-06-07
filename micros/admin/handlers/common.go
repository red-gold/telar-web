package handlers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alexellis/hmac"
	"github.com/gofiber/fiber/v2"
	coreConfig "github.com/red-gold/telar-core/config"
	"github.com/red-gold/telar-core/types"
	ac "github.com/red-gold/telar-web/micros/admin/config"

	utils "github.com/red-gold/telar-core/utils"
)

// functionCall send request to another function/microservice using HMAC validation
func functionCall(bytesReq []byte, url, method string) ([]byte, error) {
	prettyURL := utils.GetPrettyURLf(url)
	bodyReader := bytes.NewBuffer(bytesReq)
	uri := fmt.Sprintf("%s%s", *coreConfig.AppConfig.InternalGateway, prettyURL)
	fmt.Printf("\n[INFO] Function call URI [%s]", uri)

	httpReq, httpErr := http.NewRequest(method, uri, bodyReader)
	if httpErr != nil {
		return nil, httpErr
	}

	payloadSecret := *coreConfig.AppConfig.PayloadSecret

	digest := hmac.Sign(bytesReq, []byte(payloadSecret))
	httpReq.Header.Set("Content-type", "application/json")
	fmt.Printf("\ndigest: %s, header: %v \n", "sha1="+hex.EncodeToString(digest), types.HeaderHMACAuthenticate)
	httpReq.Header.Add(types.HeaderHMACAuthenticate, "sha1="+hex.EncodeToString(digest))

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
		return nil, fmt.Errorf("failed to call %s api, invalid status: %s", prettyURL, res.Status)
	}

	return resData, nil
}

// writeTokenOnCookie wite session on cookie
func writeSessionOnCookie(c *fiber.Ctx, session string, config *ac.Configuration) {
	appConfig := coreConfig.AppConfig
	parts := strings.Split(session, ".")
	headerCookie := &fiber.Cookie{
		HTTPOnly: true,
		Name:     *appConfig.HeaderCookieName,
		Value:    parts[0],
		Path:     "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	}

	payloadCookie := &fiber.Cookie{
		// HttpOnly: true,
		Name:  *appConfig.PayloadCookieName,
		Value: parts[1],
		Path:  "/",
		// Expires: time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	}

	signCookie := &fiber.Cookie{
		HTTPOnly: true,
		Name:     *appConfig.SignatureCookieName,
		Value:    parts[2],
		Path:     "/",
		// Expires:  time.Now().Add(config.CookieExpiresIn),
		Domain: config.CookieRootDomain,
	}
	// Set cookie
	c.Cookie(headerCookie)
	c.Cookie(payloadCookie)
	c.Cookie(signCookie)
}
