package handlers

import (
	_ "encoding/json"
	"github.com/cihub/seelog"
	"github.com/gorilla/websocket"
	_ "io/ioutil"
	"net/http"
	"sync"
	"tracker/models"
)

var (
	upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
)

type WebSocketHandler struct {
	BaseHandler
	Hub *models.Hub
}

func (this *WebSocketHandler) SetHub(hub *models.Hub) {
	this.Hub = hub
}
func (this *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {

	session, _ := this.Context.Get(r, "user")
	storedGPlusID := session.Values["gplusID"]
	if storedGPlusID == nil {
		seelog.Info("Gplus ID is nil for request %p", r)
		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	seelog.Infof("Upgraded connection, %p", wsConn)
	if err != nil {
		seelog.Infof("Error upgrading %s", err)
		return
	}

	c := &models.Connection{
		Send:            make(chan []byte, 256),
		Hub:             this.Hub,
		CreationTime:    GetCreationTime(r),
		Unregister:      make(chan struct{}),
		CloseConnection: make(chan struct{}),
	}

	uc := c.Hub.AddConnection(c, session)

	// defer func(uc *models.UserConnection) {
	// 	err = wsConn.Close()
	// 	if err != nil {
	// 		seelog.Infof("Cannot close socket %s", err)
	// 	}
	// 	seelog.Infof("Close connection now for %p", c)
	// 	c.Hub.RemoveConnection(uc)
	// }(uc)

	go func(uc *models.UserConnection) {
		seelog.Info("Wait for unregister signal")
		<-c.Unregister
		ws_err := wsConn.Close()
		if ws_err != nil {
			seelog.Infof("Cannot close socket %s", ws_err)
		}
		seelog.Infof("Close connection now for %p", c)
		c.Hub.RemoveConnection(uc)
		c.CloseConnection <- struct{}{}
	}(uc)

	var wg sync.WaitGroup
	// wg.Add(1) //Close reader that mean we should close Writer, don't need to Add(2) here
	go c.Reader(&wg, wsConn, storedGPlusID.(string))
	go c.Writer(&wg, wsConn)
	// wg.Wait()
	<-c.CloseConnection
}
