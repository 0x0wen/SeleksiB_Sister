package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// flownya adalah berikut:
// 1. server idup
// 2. client pertama iduo
// 3. client pertama buat akun terus login
// 4. client kedua buat aku terus login
// 5. client pertama masukin nama username terus kirim ke server
// 6. server cari di database ada ga username nya, terus balikin ip address nya (kalo ada)
// 7. client pertama mencoba konek ke ip address yang didapat dari server
// 8. kalau berhasil, keduanya masuk state chattingan

// di state chattingan:
// 1. saat pertama kali masuk state chattingan, server kirim key ke kedua client
// 2. pesan akan di encode di client dengan key yang didapat dari server
// 3. pesan akan di decode di client juga dengan key yang didapat dari server
// 4. setelah pesan sampai ke penerima, server akan mengirim key baru ke kedua client

var (
	// User database stores username to password mappings
	userDatabase = make(map[string]string) // username -> password
	// IP database stores username to connection mappings
	ipDatabase = make(map[string]string) // username -> ip
	// Mutex to protect access to the user and connection databases
	mutex       = &sync.Mutex{}
	portCounter = 2
)

const ADDRESS = "127.0.0.2"

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection error:", err)
			return
		}
		message = strings.TrimSpace(message)
		parts := strings.Split(message, " ")

		switch parts[0] {
		case "REGISTER":
			if len(parts) != 3 {
				conn.Write([]byte("INVALID_NO_OF_ARGUMENTS\n"))
				continue
			}
			username, password := parts[1], parts[2]
			registerUser(username, password, conn)
		case "LOGIN":
			if len(parts) != 3 {
				conn.Write([]byte("INVALID_NO_OF_ARGUMENTS\n"))
				continue
			}
			username, password := parts[1], parts[2]
			loginUser(username, password, conn)
		case "FIND":
			if len(parts) != 2 {
				conn.Write([]byte("INVALID_NO_OF_ARGUMENTS\n"))
				continue
			}
			username := parts[1]
			findUser(username, conn)
		default:
			conn.Write([]byte("INVALID_COMMAND\n"))
		}
	}
}

// Register a new user
func registerUser(username, password string, conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := userDatabase[username]; exists {
		conn.Write([]byte("FAILED\n"))
		return
	}

	userDatabase[username] = password
	ip := fmt.Sprintf("%s:900%d", ADDRESS, portCounter)
	portCounter++
	ipDatabase[username] = ip
	conn.Write([]byte("OK\n"))
}

// Log in an existing user
func loginUser(username, password string, conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	storedPassword, exists := userDatabase[username]
	if !exists || storedPassword != password {
		conn.Write([]byte("FAILED NOT_EXISTS\n"))
		return
	}
	conn.Write([]byte(fmt.Sprintf("OK %s\n", ipDatabase[username])))
}

// Find the IP address of a user
func findUser(username string, conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	ip, exists := ipDatabase[username]
	if exists {
		conn.Write([]byte(fmt.Sprintf("OK %s\n", ip)))
	} else {
		conn.Write([]byte("FAILED NOT_EXISTS\n"))
	}
}

func main() {
	// listen for incoming connections
	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is running on 127.0.0.1:9000")

	for {
		// accept incoming connections from clients
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		// handle each connection on separate goroutines
		go handleConnection(conn)
	}
}
