package models

import (
	"encoding/json"
	"github.com/cihub/seelog"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Connection struct {
	// Buffered channel of outbound messages.
	Send         chan []byte
	Room         string
	Hub          *Hub
	CreationTime int
}

func (c *Connection) Reader(wg *sync.WaitGroup, wsConn *websocket.Conn, userId string) {
	defer wg.Done()
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			seelog.Infof("Cannot read message from %p, error: %s\n", c, err)
			break
		}

		data := &IncommingMessage{}
		seelog.Infof("%p send: %v\n", c, string(message))
		json_err := json.Unmarshal(message, &data)

		if json_err != nil {
			data = &IncommingMessage{Room: "", Message: "", UserId: "", Cmd: 0}
		}
		data.UserId = userId
		seelog.Infof("Read Message: %v, content: %v", time.Now().UnixNano(), data)
		c.Hub.broadcast <- data
	}
}

func (c *Connection) Writer(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	for message := range c.Send {
		seelog.Infof("Write Message: %v, content: %v", time.Now().UnixNano(), string(message))
		err := wsConn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			seelog.Infof("Cannot send message %s to %p, error: %s\n", string(message), c, err)
			break
		}
	}
}
