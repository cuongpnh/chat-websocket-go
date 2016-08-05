package models

import (
	"encoding/json"
	"github.com/cihub/seelog"
	"github.com/cuongpnh/chat-websocket-go/constants"
	"github.com/gorilla/sessions"
	"strconv"
	"sync"
	"time"
)

var (
	supportedRooms = []string{"default", "18plus", "hentai"}
)

type Hub struct {
	// the mutex to protect connections
	// connectionsMx sync.RWMutex
	sync.RWMutex

	// Registered connections.
	connections     map[string]map[*UserConnection]struct{}
	userConnections map[string]*UserConnection
	// We need a thing to hold all connections related to specific user and room, it will let us remove/close user's connection easier

	users map[string]*User

	// Inbound messages from the connections.
	broadcast chan *IncommingMessage
}

func NewHub() *Hub {
	h := &Hub{
		// connectionsMx:   sync.RWMutex{},
		connections:     make(map[string]map[*UserConnection]struct{}),
		userConnections: make(map[string]*UserConnection),
		users:           make(map[string]*User),
		broadcast:       make(chan *IncommingMessage),
	}
	// for sending messages to all clients
	go func() {
		for {
			message := <-h.broadcast
			seelog.Infof("Receive message at: %v, content: %v", time.Now().UnixNano(), message)
			h.RLock()

			if _, ok := h.connections[message.Room]; !ok {
				h.RUnlock()
				break
			}
			if message.UserId == "" {
				h.RUnlock()
				continue
			}
			outgoingMessageObj := h.buildOutgoingMessage(message)
			outgoingMessageJson, _ := json.Marshal(outgoingMessageObj)
			var wg sync.WaitGroup
			switch message.Cmd {
			case constants.MESSAGE_CMD_MESSAGE:
				for uc := range h.connections[message.Room] {
					// To prevent previous messages cannot arrive client before current message should use WaitGroup here
					wg.Add(1)
					go func(uc *UserConnection) {
						defer wg.Done()
						select {
						case uc.Connection.Send <- []byte(outgoingMessageJson):
							seelog.Infof("Message for %p at %v", uc, time.Now().UnixNano())
						// stop trying to send to this connection after trying for 1 second.
						// if we have to stop, it means that a reader died so remove the connection also.
						case <-time.After(1 * time.Second):
							seelog.Infof("Shutting down connection %s", uc)
							// Send signal to c.Send or any channel => connection
							go h.RemoveConnection(uc)
						}
					}(uc)

				}
				// Handler other cases here...
			}
			wg.Wait() //Wait until we sent message to all users in current room
			h.RUnlock()
		}
	}()
	return h
}

func (h *Hub) buildOutgoingMessage(message *IncommingMessage) *OutgoingMessage {
	user := h.users[message.UserId]

	return &OutgoingMessage{
		Room:      message.Room,
		Message:   message.Message,
		UserId:    message.UserId,
		Timestamp: message.Timestamp,
		Name:      user.Name,
		Picture:   user.Picture,
		Email:     user.Email,
	}
}

func (h *Hub) AddConnection(conn *Connection, session *sessions.Session) *UserConnection {
	h.Lock()
	defer h.Unlock()

	userId := session.Values["gplusID"].(string)
	name := session.Values["name"].(string)
	picture := session.Values["picture"].(string)
	email := session.Values["email"].(string)

	// New a User if current user is not exists
	if _, ok := h.users[userId]; !ok {
		seelog.Infof("User isn't exists %p, user: %v", conn, userId)
		connections := make(map[*Connection]struct{})
		h.users[userId] = &User{Connections: connections, Id: userId, Name: name, Picture: picture, Email: email}
	}
	// Add new connection to user
	h.users[userId].Connections[conn] = struct{}{}

	// For testing purpose, we will add user to 3 rooms below
	for _, room := range supportedRooms {
		if _, ok := h.connections[room]; !ok {
			seelog.Infof("Room isn't exists %v", room)
			h.connections[room] = make(map[*UserConnection]struct{})
		}
	}

	// Add new user connection to connections list
	userConnection := &UserConnection{conn, userId}
	// seelog.Infof("Add connection: %p for user %v", conn, conn.Room)
	for _, room := range supportedRooms {
		h.connections[room][userConnection] = struct{}{}
		keyUserConnection := userId + "_" + room + "_" + strconv.Itoa(conn.CreationTime)
		h.userConnections[keyUserConnection] = userConnection
	}

	return userConnection
}

func (h *Hub) RemoveConnection(uc *UserConnection) {
	h.Lock()
	defer h.Unlock()

	userId := uc.UserId
	conn := uc.GetConnection()

	// Remove user's connection if it's exists
	if _, ok := h.users[userId].Connections[conn]; ok {

		for _, room := range supportedRooms {
			keyUserConnection := userId + "_" + room + "_" + strconv.Itoa(conn.CreationTime)
			delete(h.connections[room], uc)
			delete(h.userConnections, keyUserConnection)
		}
		delete(h.users[userId].Connections, conn)
		close(conn.Send)

		// Remove user's also if his/her connections are empty
		if len(h.users[userId].Connections) == 0 {
			seelog.Infof("Remove connection %p, user: %v", conn, userId)
			delete(h.users, userId)
		}
	}
}

func (h *Hub) CloseAllConnections() {
	seelog.Info("Close all connections now!")
	for userId := range h.users {
		for c := range h.users[userId].Connections {
			c.Lock()
			c.Unregister <- struct{}{}
			c.SendUnregisterMessage = true
			c.Unlock()
		}
		seelog.Infof("User: %v", userId)
	}
	// for _, room := range supportedRooms {
	// 	for uc := range h.connections[room] {
	// 		connection := uc.GetConnection()
	// 		seelog.Infof("Close for connection: %p for room: %s", connection, room)
	// 		connection.Unregister <- struct{}{}
	// 		// close(conn.Send)
	// 	}
	// }
}
