package models

type OutgoingMessage struct {
	Room      string `json:"room"`
	UserId    string `json:"user_id"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	Email     string `json:"email"`
}
