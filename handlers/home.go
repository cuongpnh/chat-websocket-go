package handlers

import (
	"github.com/cihub/seelog"
	"net/http"
	"text/template"
	"tracker/models"
)

type HomeHandler struct {
	BaseHandler
}

func (this *HomeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFiles("templates/index.html"))
	session, _ := this.Context.Get(r, "user")
	seelog.Infof("Store: %p", this.Context)
	seelog.Infof("Session: %p", session)
	seelog.Infof("Session data: %s", session)
	storedToken := session.Values["accessToken"]

	seelog.Infof("Access token: %v", storedToken)

	if storedToken == nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	} else {
		accessToken := &models.AccessToken{storedToken.(string)}
		tpl.Execute(w, accessToken)
		return
	}

	tpl.Execute(w, r)
}
