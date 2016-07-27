package models

type User struct {
	Connections map[*Connection]struct{}
	Id          string
	Name        string
	Picture     string
	Email       string
}
