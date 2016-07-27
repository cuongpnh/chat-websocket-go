package models

type IncommingMessage struct {
	Room      string `json:"room"`
	UserId    string `json:"user_id"`
	Message   string `json:"message"`
	Cmd       int    `json:"cmd"`
	Timestamp int    `json:"timestamp"`
}
