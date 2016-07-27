package handlers

import (
	"encoding/json"
	"github.com/cihub/seelog"
	"net/http"
)

type ReconnectHandler struct {
	BaseHandler
}

func (this *ReconnectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	session, _ := this.Context.Get(r, "user")
	accessToken := GetAccessToken(r)

	result := make(map[string]bool)
	result["success"] = false
	resultJs, err := json.Marshal(result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if accessToken == "" {
		w.Write(resultJs)
		return
	}
	data, err := this.GetUserInfo(accessToken)
	if err != nil {
		seelog.Infof("Error when try to get user info: %s", err)
		w.Write(resultJs)
		return
	}
	// response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	// contents, err := ioutil.ReadAll(response.Body)

	// if err != nil {
	// 	w.Write(resultJs)
	// 	return
	// }
	// data := &models.UserGoogle{}
	// json_err := json.Unmarshal(contents, &data)

	// if json_err != nil {
	// 	w.Write(resultJs)
	// 	return
	// }
	// if data.Id == "" {
	// 	w.Write(resultJs)
	// 	return
	// }
	// seelog.Infof("%s", data)
	// Store user info
	data.AccessToken = accessToken
	this.UpdateUserSession(session, data, w, r)
	seelog.Infof("Reconnect Access token: %v", accessToken)

	result["success"] = true
	resultJs, _ = json.Marshal(result)
	w.Write(resultJs)
	return
}
