package handlers

import (
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
)

const cacheTimeout = 7200

func generateV4GetObjectSignedURL(bucketName string, objectName string, serviceAccount string) (string, error) {
	// [START storage_generate_signed_url_v4]

	conf, err := google.JWTConfigFromJSON([]byte(serviceAccount))
	if err != nil {
		return "", fmt.Errorf("google.JWTConfigFromJSON: %v", err)
	}

	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Expires:        time.Now().Add(cacheTimeout * time.Second),
	}

	u, err := storage.SignedURL(bucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("Unable to generate a signed URL: %v", err)
	}

	fmt.Println("Generated GET signed URL:")
	fmt.Printf("%q\n", u)
	// [END storage_generate_signed_url_v4]
	return u, nil
}
