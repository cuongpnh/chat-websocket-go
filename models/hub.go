package models

import (
	"encoding/json"
	"github.com/cihub/seelog"
	"github.com/gorilla/sessions"
	"go-in-5-minutes/episode4/constants"
	"strconv"
	"sync"
	"time"
)

type Hub struct {
	// the mutex to protect connections
	connectionsMx sync.RWMutex
	logMx         sync.RWMutex

	// Registered connections.
	connections     map[string]map[*UserConnection]struct{} //[room][user_connection]
	userConnections map[string]*UserConnection
	// We need a thing to hold all connections related to specific user and room, it will let us remove/close user's connection easier

	users map[string]*User

	// Inbound messages from the connections.
	broadcast chan *IncommingMessage

	log [][]byte
}

func NewHub() *Hub {
	// users: make(map[int]*User)
	h := &Hub{
		connectionsMx:   sync.RWMutex{},
		connections:     make(map[string]map[*UserConnection]struct{}),
		userConnections: make(map[string]*UserConnection),
		users:           make(map[string]*User),
		broadcast:       make(chan *IncommingMessage),
	}
	// for sending messages to all clients
	go func() {
		for {
			message := <-h.broadcast
			seelog.Infof("%v", message)
			h.connectionsMx.RLock()

			if _, ok := h.connections[message.Room]; !ok {
				h.connectionsMx.RUnlock()
				break
			}
			if message.UserId == "" {
				h.connectionsMx.RUnlock()
				continue
			}

			user := h.users[message.UserId]

			outgoingMessageObj := &OutgoingMessage{
				Room:      message.Room,
				Message:   message.Message,
				UserId:    message.UserId,
				Timestamp: message.Timestamp,
				Name:      user.Name,
				Picture:   user.Picture,
				Email:     user.Email,
			}
			outgoingMessageJson, _ := json.Marshal(outgoingMessageObj)

			switch message.Cmd {
			case constants.MESSAGE_CMD_MESSAGE:
				for c := range h.connections[message.Room] {
					// Use goroutine here
					select {
					case c.Connection.Send <- []byte(outgoingMessageJson):
						seelog.Infof("Message for %p at %v", c, time.Now().UnixNano())
					// stop trying to send to this connection after trying for 1 second.
					// if we have to stop, it means that a reader died so remove the connection also.
					case <-time.After(1 * time.Second):
						seelog.Infof("Shutting down connection %s", c)
						// Send signal to c.Send or any channel => connection
						go h.RemoveConnection(c)
					}
				}
			case constants.MESSAGE_CMD_CLOSE:
				keyUserConnection := message.UserId + "_" + message.Room + "_" + strconv.Itoa(message.CreationTime)
				userConnection := h.userConnections[keyUserConnection]
				go h.RemoveConnection(userConnection)
			case constants.MESSAGE_CMD_LEAVE:
			}

			h.connectionsMx.RUnlock()
		}
	}()
	return h
}

func (h *Hub) AddConnection(conn *Connection, session *sessions.Session) *UserConnection {
	h.connectionsMx.Lock()
	defer h.connectionsMx.Unlock()

	userId := session.Values["gplusID"].(string)
	name := session.Values["name"].(string)
	picture := session.Values["picture"].(string)
	email := session.Values["email"].(string)

	// New a User if current user is not exists
	if _, ok := h.users[userId]; !ok {
		seelog.Infof("User isn't exists %p, user: %v, room: %v", conn, userId, conn.Room)
		connections := make(map[*Connection]struct{})
		h.users[userId] = &User{Connections: connections, Id: userId, Name: name, Picture: picture, Email: email}
	}
	// Add new connection to user
	h.users[userId].Connections[conn] = struct{}{}

	// For testing purpose, we will add user to 3 rooms below
	rooms := []string{"default", "18plus", "hentai"}
	for _, room := range rooms {
		if _, ok := h.connections[room]; !ok {
			seelog.Infof("Room isn't exists %v", room)
			h.connections[room] = make(map[*UserConnection]struct{})
		}
	}

	// if _, ok := h.connections[conn.Room]; !ok {
	// 	seelog.Infof("Room isn't exists %v", conn.Room)
	// 	h.connections[conn.Room] = make(map[*UserConnection]struct{})
	// }

	// Add new user connection to connections list
	userConnection := &UserConnection{conn, userId}
	// seelog.Infof("Add connection: %p for user %v", conn, conn.Room)
	for _, room := range rooms {
		h.connections[room][userConnection] = struct{}{}
		keyUserConnection := userId + "_" + room + "_" + strconv.Itoa(conn.CreationTime)
		h.userConnections[keyUserConnection] = userConnection
	}
	// h.connections[conn.Room][userConnection] = struct{}{}
	// keyUserConnection := strconv.Itoa(userId) + "_" + conn.Room
	// h.userConnections[keyUserConnection] = userConnection
	return userConnection
}

func (h *Hub) RemoveConnection(uc *UserConnection) {
	h.connectionsMx.Lock()
	defer h.connectionsMx.Unlock()

	userId := uc.UserId
	conn := uc.GetConnection()

	// Remove user's connection if it's exists
	if _, ok := h.users[userId].Connections[conn]; ok {
		rooms := []string{"default", "18plus", "hentai"}
		for _, room := range rooms {
			keyUserConnection := userId + "_" + room + "_" + strconv.Itoa(conn.CreationTime)
			delete(h.connections[room], uc)
			delete(h.userConnections, keyUserConnection)
		}
		// keyUserConnection := strconv.Itoa(userId) + "_" + conn.Room + "_" + strconv.Itoa(conn.CreationTime)
		// delete(h.connections[conn.Room], uc)
		// delete(h.userConnections, keyUserConnection)
		delete(h.users[userId].Connections, conn)
		close(conn.Send)

		// Remove user's also if his/her connections are empty
		if len(h.users[userId].Connections) == 0 {
			seelog.Infof("Remove connection %p, user: %v", conn, userId)
			delete(h.users, userId)
		}
	}
}
