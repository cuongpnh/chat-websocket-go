package main

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/cuongpnh/chat-websocket-go/env"
	"github.com/cuongpnh/chat-websocket-go/handlers"
	"github.com/cuongpnh/chat-websocket-go/models"
	"github.com/fvbock/endless"
	"net/http"
	"syscall"
)

func main() {

	defer seelog.Flush()
	logConfig := env.Get("LOG_CONFIG")
	logger, err := seelog.LoggerFromConfigAsFile(logConfig)
	if err != nil {
		panic(err)
	}
	seelog.ReplaceLogger(logger)

	h := models.NewHub()

	seelog.Infof("Hub: %p", h)
	router := http.NewServeMux()
	router.HandleFunc("/", handlers.NewHandler(&handlers.HomeHandler{}))
	router.HandleFunc("/login", handlers.NewHandler(&handlers.LoginGoogleHandler{}))
	router.HandleFunc("/callback", handlers.NewHandler(&handlers.LoginGoogleCallbackHandler{}))
	router.HandleFunc("/reconnect", handlers.NewHandler(&handlers.ReconnectHandler{}))
	router.HandleFunc("/ws", handlers.NewHandler(&handlers.WebSocketHandler{Hub: h}))

	url := fmt.Sprintf(":%s", env.Get("PORT"))
	seelog.Info("Serving on port " + url)

	endlessServer := endless.NewServer(url, router)
	endlessServer.SignalHooks[endless.PRE_SIGNAL][syscall.SIGINT] = append(
		endlessServer.SignalHooks[endless.PRE_SIGNAL][syscall.SIGINT],
		h.CloseAllConnections)

	seelog.Error(endlessServer.ListenAndServe())
}

// Flow
// 1. Client send messages via UI by using `send`
// 2. Its socket caughts message by using `wsConn.ReadMessage()` then send message to hub's broadcast
// 3. Broadcast takes message and sent it to all active clients (send message to channel `send` of connection))
// 4. Client caughts message by listening channel `send` of connection
// 5. Write message to its socket
// 6. UI takes message via callback `onmessage`
