package handlers

import (
	"github.com/cihub/seelog"
	"github.com/cuongpnh/chat-websocket-go/models"
	"github.com/gorilla/websocket"
	"net/http"
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

	go c.Reader(wsConn, storedGPlusID.(string))
	go c.Writer(wsConn)

	<-c.CloseConnection
}
