package utils

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     "491978716561-8jlo7n7mkpmdfi6mdrqe62a39b1gs9st.apps.googleusercontent.com", //os.Getenv("GOOGLE_KEY"),
		ClientSecret: "Nx2Qt89DjVqjZZc-IMnP2wPk",                                                 //os.Getenv("GOOGLE_SECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
}
