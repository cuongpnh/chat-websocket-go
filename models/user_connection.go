package models

type UserConnection struct {
	Connection *Connection
	UserId     string
}

func (uc *UserConnection) GetConnection() *Connection {
	return uc.Connection
}
