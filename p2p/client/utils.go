package main

import (
	"fmt"
	"os"
)

type Authentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChatMessage struct {
    Sender    string
    Receiver  string
    Message   string
}

type Message struct {
	Key        []byte `json:"key"`
	Ciphertext []byte `json:"ciphertext"`
}

type Body struct {
	Type    string
	Message string
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
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
func (a Authentication) toString() string {
	return a.Username + " " + a.Password
}

func (m Message) toString() string {
	return string(m.Key) + " " + string(m.Ciphertext)
}
