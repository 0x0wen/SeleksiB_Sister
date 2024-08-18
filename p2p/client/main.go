package main

import (
	"bufio"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	privateKey, _ = getOrCreateRSAKey("private_key.txt") // Generate a pair of asymmetric keys
	publicKey       = &privateKey.PublicKey
	AESKey,_ 		 = getOrCreateAESKey("aes_key.txt")
	myAddress       = ""
	myUsername      string
	peer            net.Conn
	peerUsername    string
	peerPublicKey   *rsa.PublicKey
	messageLabel = widget.NewLabel("")
	messageList       *widget.List	
	messageContainer  *fyne.Container
	messages            []ChatMessage
	messagesMutex     = &sync.Mutex{}
	window fyne.Window
)

// things to do:
// 1. use handleerr for all errors

const serverAddress = "127.0.0.1:9000"

func main() {
	initDB()
	defer db.Close()
    myApp := app.New()
    window=  myApp.NewWindow("Login/Register")
    // Login form
    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Username")
    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Password")
    loginForm := widget.NewForm(
        widget.NewFormItem("Username", usernameEntry),
        widget.NewFormItem("Password", passwordEntry),
    )
    loginForm.OnSubmit = func() {
		status, ip := login(usernameEntry.Text, passwordEntry.Text)
		if status {
			messageLabel.SetText("Login successful!")
			myUsername = usernameEntry.Text
			myAddress = ip
			go startListening()
			showPeerEntryForm(window)
		} else {
			messageLabel.SetText("Wrong credentials! Login failed.")
		}
        fmt.Println("Login Submitted")
        fmt.Println("Username:", usernameEntry.Text)
        fmt.Println("Password:", passwordEntry.Text)
    }

    // Register form
    regUsernameEntry := widget.NewEntry()
    regUsernameEntry.SetPlaceHolder("Username")
    regPasswordEntry := widget.NewPasswordEntry()
    regPasswordEntry.SetPlaceHolder("Password")
    registerForm := widget.NewForm(
        widget.NewFormItem("Username", regUsernameEntry),
        widget.NewFormItem("Password", regPasswordEntry),
    )
    registerForm.OnSubmit = func() {
		if register(regUsernameEntry.Text, regPasswordEntry.Text) {
			messageLabel.SetText("Register successful!")
			} else {
			messageLabel.SetText("Username taken! Register failed.")
		}
        fmt.Println("Register Submitted")
        fmt.Println("Username:", regUsernameEntry.Text)
        fmt.Println("Password:", regPasswordEntry.Text)
    }

    // Tabs for Login and Register
    tabs := container.NewAppTabs(
        container.NewTabItem("Login", loginForm),
        container.NewTabItem("Register", registerForm),
    )
	content := container.NewVBox(
		messageLabel,
		tabs,
	)
    window.SetContent(content)
    window.Resize(fyne.NewSize(400, 300))
    window.ShowAndRun()
}



func showPeerEntryForm(window fyne.Window) {
	peerUsernameEntry := widget.NewEntry()
	peerUsernameEntry.SetPlaceHolder("Peer Username")
	peerEntryForm := widget.NewForm(
		widget.NewFormItem("Peer Username", peerUsernameEntry),
	)
	peerEntryForm.OnSubmit = func() {
		fmt.Println("Peer Username Submitted:", peerUsernameEntry.Text)
		peerUsername = peerUsernameEntry.Text
		status, peerIP := find(peerUsername)
		if status {
			peer = createConnection(peerIP)
			go func() {
				err := sendToPeer(fmt.Sprintf("CONNECT %s %s\n", myUsername, myAddress))
				handleErr(err)
				err = sendToPeer(fmt.Sprintf("PUBLIC_KEY %s %d\n", publicKey.N.Text(16), publicKey.E))
				handleErr(err)
				messages = retrieveChatHistory(myUsername, peerUsername)
			}()
			fmt.Print("Connected to ", peerUsername, "\n")
			go receiveMessage(peer)
			showChatWindow(window)
			peerUsernameEntry.SetText("")
		} else {
			messageLabel.SetText("Peer not found!")
		}
	}

	window.SetContent(container.NewVBox(
		messageLabel,
		peerEntryForm,
		widget.NewButton("Back to Login", func() {
			showLoginForm(window)
		}),
	))
}
func showLoginForm(window fyne.Window) *fyne.Container {
    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Username")
    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Password")

    form := &widget.Form{
        Items: []*widget.FormItem{
            {Text: "Username", Widget: usernameEntry},
            {Text: "Password", Widget: passwordEntry},
        },
        OnSubmit: func() {
            // Handle login logic here
            username := usernameEntry.Text
            password := passwordEntry.Text
            fyne.LogError("Login submitted: "+username+" / "+password, nil)
        },
    }

    return container.NewVBox(
		messageLabel,
        form,
        widget.NewButton("Login", func() {
            form.OnSubmit()
        }),
        widget.NewButton("Back to Login", func() {
            window.SetContent(container.NewVBox(
                showLoginForm(window),
                widget.NewButton("Register", func() {
                    window.SetContent(showRegisterForm(window))
                }),
            ))
        }),   

    )
}

func addMessage(sender string,receiver string, message string) {
	messagesMutex.Lock()
	defer messagesMutex.Unlock()
	chatMessage := ChatMessage{
		Sender:   sender,
		Receiver: receiver,
		Message:  message,
	}
	messages = append(messages, chatMessage)

}

func sendMessage(message string){
	key, cipherText, err := OTPEncrypt([]byte(message))
	handleErr(err)
	cipherKey, err := EncryptOAEP([]byte(key), nil, peerPublicKey)
	handleErr(err)
	encodedEncryptedKey := base64.StdEncoding.EncodeToString(cipherKey)
	encodedCipherText := base64.StdEncoding.EncodeToString(cipherText)
	go func(){
		err := sendToPeer(fmt.Sprintf("MESSAGE %s %s\n", encodedEncryptedKey, encodedCipherText))
		handleErr(err)
	}()
	addMessage(myUsername, peerUsername, message)
	messageList.ScrollToBottom()
	go storeMessage(myUsername, peerUsername, message)
}

func showChatWindow(window fyne.Window) {
	chatInput := widget.NewEntry()
	chatInput.SetPlaceHolder("Enter your message")
	// on enter
	chatInput.OnSubmitted = func(message string) {
		sendMessage(message)
		chatInput.SetText("")
	}

	// on send button clicked
	sendButton := widget.NewButton("Send", func() {
		message := chatInput.Text
		sendMessage(message)
		chatInput.SetText("")
	})
	// Create a custom widget for chat messages
	createMessageBubble := func(sender, message string) fyne.CanvasObject {
		var senderLabel *widget.Label
		var messageLabel *widget.Label
		if(sender == myUsername){
			senderLabel = widget.NewLabelWithStyle(sender, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
			messageLabel = widget.NewLabelWithStyle(message, fyne.TextAlignTrailing, fyne.TextStyle{Bold: false})	
		}else{
			senderLabel = widget.NewLabelWithStyle(sender, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			messageLabel = widget.NewLabelWithStyle(message, fyne.TextAlignLeading, fyne.TextStyle{Bold: false})	
		}
		messageLabel.Wrapping = fyne.TextWrapWord
		messageBubble := container.NewVBox(senderLabel, messageLabel)
		return messageBubble
	}

	messageList = widget.NewList(
		func() int {
			messagesMutex.Lock()
			defer messagesMutex.Unlock()
			return len(messages)
		},
		func() fyne.CanvasObject {
			// Return the message bubble directly
			return createMessageBubble("", "")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			messagesMutex.Lock()
			defer messagesMutex.Unlock()
			message := messages[i]
			
			// Ensure that 'o' is correctly handled
			messageContainer := o.(*fyne.Container)
			messageContainer.Objects = []fyne.CanvasObject{createMessageBubble(message.Sender, message.Message)}
			messageContainer.Refresh()
		},
	)

	messageList.Resize(fyne.NewSize(500, 450))
	messageList.ScrollToBottom()
	inputSection := container.NewWithoutLayout(chatInput, sendButton)
	chatInput.Resize(fyne.NewSize(400, 50))
	chatInput.Move(fyne.NewPos(0, 0))
	sendButton.Resize(fyne.NewSize(100, 50))
	sendButton.Move(fyne.NewPos(400, 0))
	inputSection.Move(fyne.NewPos(0, 450))
	inputSection.Resize(fyne.NewSize(500, 50))
	messageContainer := container.NewWithoutLayout(
		messageList,
		inputSection,
	)

	window.SetContent(messageContainer)
	window.Resize(fyne.NewSize(500, 500))
}


func showRegisterForm(window fyne.Window) *fyne.Container {
    usernameEntry := widget.NewEntry()
    usernameEntry.SetPlaceHolder("Username")
    passwordEntry := widget.NewPasswordEntry()
    passwordEntry.SetPlaceHolder("Password")
    confirmPasswordEntry := widget.NewPasswordEntry()
    confirmPasswordEntry.SetPlaceHolder("Confirm Password")

    form := &widget.Form{
        Items: []*widget.FormItem{
            {Text: "Username", Widget: usernameEntry},
            {Text: "Password", Widget: passwordEntry},
            {Text: "Confirm Password", Widget: confirmPasswordEntry},
        },
        OnSubmit: func() {
            // Handle registration logic here
            username := usernameEntry.Text
            password := passwordEntry.Text
            confirmPassword := confirmPasswordEntry.Text
            if password != confirmPassword {
                fyne.LogError("Passwords do not match", nil)
            } else {
                fyne.LogError("Registration submitted: "+username+" / "+password, nil)
            }
        },
    }

    return container.NewVBox(
		messageLabel,
        form,
        widget.NewButton("Register", func() {
            form.OnSubmit()
        }),
        widget.NewButton("Back to Login", func() {
            window.SetContent(container.NewVBox(
                showLoginForm(window),
                widget.NewButton("Register", func() {
                    window.SetContent(showRegisterForm(window))
                }),
            ))
        }),
    )
}

func connectToServer() (net.Conn, error) {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	return conn, nil
}

func sendToServer(request string) (string, error) {
	conn, err := connectToServer()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	fmt.Fprintf(conn, request+"\n")
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(response), nil
}

func sendToPeer(request string) error {
	if peer == nil {
		return fmt.Errorf("peer not connected")
	}
	fmt.Fprintf(peer, request+"\n")
	return nil
}

func register(username, password string) bool {
	response, err := sendToServer(fmt.Sprintf("REGISTER %s %s", username, password))
	parts := strings.Split(strings.TrimSpace(response), " ")
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	return parts[0] == "OK"
}

func login(username, password string) (bool, string) {
	response, err := sendToServer(fmt.Sprintf("LOGIN %s %s", username, password))
	parts := strings.Split(strings.TrimSpace(response), " ")
	if err != nil {
		fmt.Println("Error:", err)
		return false, ""
	}
	return parts[0] == "OK", parts[1]
}

func find(username string) (bool, string) {
	response, err := sendToServer(fmt.Sprintf("FIND %s", username))
	parts := strings.Split(strings.TrimSpace(response), " ")
	if err != nil {
		fmt.Println("Error:", err)
		return false, ""
	}
	return parts[0] == "OK", parts[1]
}

func handleErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func createConnection(IP string) net.Conn {
	service := IP
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	handleErr(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	handleErr(err)
	return conn
}

func startListening() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", myAddress)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go receiveMessage(conn)
	}
}


func receiveMessage(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}
		parts := strings.SplitN(message, " ", 2)
		content_type := parts[0]
		content := strings.Split(parts[1], " ")
		switch content_type { // the body type
		case "CONNECT":
			// clearScreen()
			peerUsername = content[0]
			peerAddress := content[1]
			peer = createConnection(peerAddress)
			// send public key to peer
			go func(){
				err := sendToPeer(fmt.Sprintf("PUBLIC_KEY %s %d\n", publicKey.N.Text(16), publicKey.E))
				handleErr(err)
			}()
			fmt.Print("You are chatting with ", peerUsername, "\n")
		case "PUBLIC_KEY":
			modulus := new(big.Int)
			modulus.SetString(content[0], 16) // Base 16 (hexadecimal)
			exponent := new(big.Int)
			exponent.SetString(content[1], 10)
			peerPublicKey = &rsa.PublicKey{
				N: modulus,
				E: int(exponent.Int64()),
			}
			messages = retrieveChatHistory(myUsername, peerUsername)
			showChatWindow(window)
		case "MESSAGE":
			decodedEncryptedKey, _ := base64.StdEncoding.DecodeString(content[0])
			key, err := DecryptOAEP(decodedEncryptedKey , nil, privateKey)
			handleErr(err)
			decodedCipherText, _ := base64.StdEncoding.DecodeString(content[1])
			message := OTPDecrypt(key, decodedCipherText)
			fmt.Printf("%s > %s\n", peerUsername, message)
			go storeMessage(peerUsername, myUsername, string(message))
			addMessage( peerUsername,myUsername, string(message))
			messageList.ScrollToBottom()
		}
	}
}
