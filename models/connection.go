package models

import (
	"encoding/json"
	"github.com/cihub/seelog"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Connection struct {
	sync.RWMutex
	// Buffered channel of outbound messages.
	Send                  chan []byte
	Hub                   *Hub
	CreationTime          int
	Unregister            chan struct{}
	CloseConnection       chan struct{}
	SendUnregisterMessage bool
}

func (c *Connection) Reader(wg *sync.WaitGroup, wsConn *websocket.Conn, userId string) {
	// wg.Done()
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			seelog.Infof("Cannot read message from %p, error: %s\n", c, err)
			if c.SendUnregisterMessage == false {
				c.Lock()
				c.Unregister <- struct{}{}
				c.SendUnregisterMessage = true
				c.Unlock()
			}
			break
		}

		data := &IncommingMessage{}
		seelog.Infof("%p send: %v\n", c, string(message))
		json_err := json.Unmarshal(message, &data)

		if json_err != nil {
			data = &IncommingMessage{Message: "", UserId: "", Cmd: 0}
		}
		data.UserId = userId
		seelog.Infof("Read Message: %v, content: %v", time.Now().UnixNano(), data)
		c.Hub.broadcast <- data
	}
	seelog.Infof("Close read channel for %p", c)
}

func (c *Connection) Writer(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	for message := range c.Send {
		seelog.Infof("Write Message for %p, at : %v, content: %v", c, time.Now().UnixNano(), string(message))
		err := wsConn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			seelog.Infof("Cannot send message %s to %p, error: %s\n", string(message), c, err)
			break
		}
	}
	seelog.Infof("Close write channel for %p", c)
}
