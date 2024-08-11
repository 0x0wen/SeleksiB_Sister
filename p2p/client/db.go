package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./chat_history.db")
	handleErr(err)

	// Create chat_history table if not exists
	createTableSQL := `CREATE TABLE IF NOT EXISTS chat_history (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"sender" TEXT,
		"receiver" TEXT,
		"message" TEXT,
		"timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(createTableSQL)
	handleErr(err)
}

func storeMessage(sender, receiver, message string) {
	insertSQL := `INSERT INTO chat_history (sender, receiver, message) VALUES (?, ?, ?)`
	encryptedMessage,err := encryptAES([]byte(AESKey),[]byte(message))
	handleErr(err)
	_, err = db.Exec(insertSQL, sender, receiver, encryptedMessage)
	handleErr(err)
}

func retrieveChatHistory(sender, receiver string) []ChatMessage{
	query := `SELECT sender,receiver, message FROM chat_history WHERE (sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?) ORDER BY timestamp`
	rows, err := db.Query(query, sender, receiver, receiver, sender)
	handleErr(err)
	defer rows.Close()

	messages := []ChatMessage{}
	for rows.Next() {
		var msgSender,msgReceiver, message string
		err = rows.Scan(&msgSender, &msgReceiver,&message)
		handleErr(err)
		decryptedMessage,err := decryptAES([]byte(AESKey),message)
		handleErr(err)
		messages = append(messages, ChatMessage{Sender: msgSender, Receiver: msgReceiver, Message: decryptedMessage})
	}
	return messages
}