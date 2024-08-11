package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"time"
)

var(
	
	isWaitingPeer bool

)
func waitForPeer() {
	// Create a ticker that triggers every 300 milliseconds
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	// Create a slice of strings representing the loading animation
	loadingAnimation := []string{"Loading.", "Loading..", "Loading..."}

	for isWaitingPeer {
		for _, frame := range loadingAnimation {
			fmt.Printf("\r%s", frame) // Print the current frame
			<-ticker.C                // Wait for the next tick
		}
	}
	inputMessage()
}

func handleChat() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("1. Chat with a user")
	fmt.Println("2. Wait for a user to chat with you")
	fmt.Print("What do you want to do? (1/2): ")
	scanner.Scan()
	if scanner.Text() == "1" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter the username: ")
		scanner.Scan()
		username := scanner.Text()
		status, peer_ip := find(username)
		if !status {
			return
		}
		peer = createConnection(peer_ip)
		peerUsername = username
		err := sendToPeer(fmt.Sprintf("CONNECT %s %s\n", myUsername, myAddress))
		handleErr(err)
		err = sendToPeer(fmt.Sprintf("PUBLIC_KEY %s %d\n", publicKey.N.Text(16), publicKey.E))
		handleErr(err)
		messages = retrieveChatHistory(myUsername, peerUsername)
		fmt.Print("Connected to ", peerUsername, "\n")
		go receiveMessage(peer)
		inputMessage()
	} else {
		isWaitingPeer = true
		waitForPeer()
	}
}

func inputMessage() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s > ", myUsername)
		scanner.Scan()
		message := scanner.Text()
		key, cipherText, err := OTPEncrypt([]byte(message))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		cipherKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, peerPublicKey, []byte(key), nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		encodedEncryptedKey := base64.StdEncoding.EncodeToString(cipherKey)
		encodedCipherText := base64.StdEncoding.EncodeToString(cipherText)
		err = sendToPeer(fmt.Sprintf("MESSAGE %s %s\n", encodedEncryptedKey, encodedCipherText))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

	}
}