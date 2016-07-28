package utils

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"tracker/env"
)

func Base64Decode(s string) ([]byte, error) {
	// add back missing padding
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

// randomString returns a random string with the specified length
func RandomString(length int) (str string) {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func GetGooleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  GetGoogleRedirectURL(),
		ClientID:     env.Get("GOOGLE_KEY"),
		ClientSecret: env.Get("GOOGLE_SECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
}
func GetGoogleRedirectURL() string {
	return env.Get("PROTOCOL") + "://" + env.Get("HOST") + ":" + env.Get("APP_PORT") + "/callback"
}
