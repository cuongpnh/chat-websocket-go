package handlers

import (
	_ "encoding/json"
	"github.com/cihub/seelog"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"go-in-5-minutes/episode4/models"
	_ "io/ioutil"
	"log"
	"net/http"
	"sync"
)

var (
	upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
)

type WsHandler struct {
	Hub     *models.Hub
	Context *sessions.CookieStore
}

func (wsh WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	session, _ := wsh.Context.Get(r, "user")
	storedGPlusID := session.Values["gplusID"]
	if storedGPlusID == nil {
		log.Println("Gplus ID is nil")
		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	seelog.Infof("Upgraded connection, %p\n", wsConn)
	if err != nil {
		seelog.Infof("Error upgrading %s", err)
		return
	}

	c := &models.Connection{
		Send:         make(chan []byte, 256),
		Hub:          wsh.Hub,
		Room:         GetRoom(r),
		CreationTime: GetCreationTime(r),
	}

	uc := c.Hub.AddConnection(c, session)

	defer func(uc *models.UserConnection) {
		err = wsConn.Close()
		if err != nil {
			seelog.Infof("Cannot close socket %s", err)
		}
		seelog.Infof("Close connection now for %p\n", c)
		c.Hub.RemoveConnection(uc)
	}(uc)

	var wg sync.WaitGroup
	wg.Add(1) //Close reader that mean we should close Writer, don't need to Add(2) here
	go c.Reader(&wg, wsConn, storedGPlusID.(string))
	go c.Writer(&wg, wsConn)
	wg.Wait()

}
