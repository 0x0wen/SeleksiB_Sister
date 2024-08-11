package main

import "fmt"

type Authentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	Key        []byte `json:"key"`
	Ciphertext []byte `json:"ciphertext"`
}

type Body struct {
	Type    string
	Message string
}
type UserPair struct {
	User1 string
	User2 string
}

func (p UserPair) Key() string {
	if p.User1 < p.User2 {
		return fmt.Sprintf("%s_%s", p.User1, p.User2)
	}
	return fmt.Sprintf("%s_%s", p.User2, p.User1)
}
func createMessage(content string) Message {
	return Message{
		Key:        []byte(""),
		Ciphertext: []byte(content),
	}
}

func createAuthentication(username string, password string) Authentication {
	return Authentication{
		Username: username,
		Password: password,
	}
}

func (a Authentication) toString() string {
	return a.Username + " " + a.Password
}

func (m Message) toString() string {
	return string(m.Key) + " " + string(m.Ciphertext)
}
