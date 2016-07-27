package handlers

import (
	"go-in-5-minutes/episode4/utils"
	"net/http"
)

type LoginGoogleHandler struct {
	BaseHandler
}

func (this *LoginGoogleHandler) Handle(w http.ResponseWriter, r *http.Request) {
	oauthStateString := utils.RandomString(64)
	session, _ := this.Context.Get(r, "user")
	session.Values["state"] = oauthStateString
	session.Save(r, w)
	url := utils.GetGooleOauthConfig().AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
