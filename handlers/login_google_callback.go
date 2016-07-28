package handlers

import (
	"github.com/cihub/seelog"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"tracker/utils"
)

type LoginGoogleCallbackHandler struct {
	BaseHandler
}

func (this *LoginGoogleCallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	session, err := this.Context.Get(r, "user")
	state := r.FormValue("state")
	if state != session.Values["state"].(string) {
		seelog.Infof("Invalid oauth state, expected '%s', got '%s'\n", session.Values["state"].(string), state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// Save code for refreshing token in future
	code := r.FormValue("code")
	token, err := utils.GetGooleOauthConfig().Exchange(oauth2.NoContext, code)
	accessToken := token.AccessToken
	if err != nil {
		log.Println("Code exchange failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Store accessToken if it isn't exists or has been changed
	storedToken := session.Values["accessToken"]
	needUpdate := false
	if storedToken == nil || (storedToken != nil && storedToken != accessToken) {
		needUpdate = true
	}
	seelog.Infof("Access Token %s\n", accessToken)

	data, err := this.GetUserInfo(accessToken)
	if err != nil {
		seelog.Infof("Error when try to get user info: %s", err)
		return
	}

	if needUpdate {
		data.AccessToken = accessToken
		this.UpdateUserSession(session, data, w, r)
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
