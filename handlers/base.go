package handlers

import (
	"encoding/json"
	"errors"
	"github.com/cihub/seelog"
	"github.com/gorilla/sessions"
	"go-in-5-minutes/episode4/models"
	"go-in-5-minutes/episode4/utils"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	sessionKey = utils.RandomString(32)
	store      = sessions.NewCookieStore([]byte(sessionKey))
)

type BaseHandler struct {
	Context *sessions.CookieStore
}

type Handler interface {
	SetContext(*sessions.CookieStore)
	Handle(http.ResponseWriter, *http.Request)
}

func (this *BaseHandler) SetContext(context *sessions.CookieStore) {
	this.Context = context
}

func NewHandler(handler Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seelog.Info("Set context")
		handler.SetContext(store)
		handler.Handle(w, r)
	})
}

func (this *BaseHandler) GetUserInfo(accessToken string) (*models.UserGoogle, error) {
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	data := &models.UserGoogle{}
	json_err := json.Unmarshal(contents, &data)

	if json_err != nil {
		return nil, json_err
	}
	if data.Id == "" {
		return nil, errors.New("Id is empty")
	}

	seelog.Infof("%s", data)
	return data, nil
}

func (this *BaseHandler) UpdateUserSession(session *sessions.Session, data *models.UserGoogle, w http.ResponseWriter, r *http.Request) {
	session.Values["email"] = data.Email
	session.Values["picture"] = data.Picture
	session.Values["name"] = data.Name
	session.Values["gplusID"] = data.Id
	session.Values["accessToken"] = data.AccessToken
	session.Save(r, w)
}

func GetRoom(r *http.Request) string {
	room := r.URL.Query().Get("room")
	if len(room) == 0 {
		room = "default"
	}
	return room
}

func GetAccessToken(r *http.Request) string {
	return r.URL.Query().Get("access_token")
}

func GetCreationTime(r *http.Request) int {
	time := r.URL.Query().Get("creation_time")
	timeInt, _ := strconv.Atoi(time)
	return timeInt
}
