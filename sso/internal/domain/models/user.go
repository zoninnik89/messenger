package models

type User struct {
	ID       string
	Login    string
	PassHash []byte
}
