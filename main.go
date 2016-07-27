package main

import (
	_ "fmt"
	"github.com/cihub/seelog"
	"github.com/gorilla/sessions"
	"go-in-5-minutes/episode4/handlers"
	"go-in-5-minutes/episode4/models"
	"go-in-5-minutes/episode4/utils"
	"net/http"
)

var (
	sessionKey = utils.RandomString(32)
	store      = sessions.NewCookieStore([]byte(sessionKey))
)

func main() {

	defer seelog.Flush()
	// logConfig := config.Get("LOG_CONFIG")
	logConfig := "log_config.xml"
	logger, err := seelog.LoggerFromConfigAsFile(logConfig)
	if err != nil {
		panic(err)
	}
	seelog.ReplaceLogger(logger)

	h := models.NewHub()
	seelog.Infof("Hub: %p", h)
	router := http.NewServeMux()
	router.HandleFunc("/", handlers.NewHandler(&handlers.HomeHandler{}, store))
	router.HandleFunc("/login", handlers.NewHandler(&handlers.LoginGoogleHandler{}, store))
	router.HandleFunc("/callback", handlers.NewHandler(&handlers.LoginGoogleCallbackHandler{}, store))
	router.HandleFunc("/reconnect", handlers.NewHandler(&handlers.ReconnectHandler{}, store))
	router.Handle("/ws", &handlers.WsHandler{Hub: h, Context: store})
	seelog.Info("Serving on port 8080")
	seelog.Error(http.ListenAndServe(":8080", router))
}

// Flow
// 1. Client send messages via UI by using `send`
// 2. Its socket caughts message by using `wsConn.ReadMessage()` then send message to hub's broadcast
// 3. Broadcast takes message and sent it to all active clients (send message to channel `send` of connection))
// 4. Client caughts message by listening channel `send` of connection
// 5. Write message to its socket
// 6. UI takes message via callback `onmessage`
