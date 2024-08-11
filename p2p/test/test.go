package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"strings"
)

func exportRSAPrivateKeyAsPEM(privateKey *rsa.PrivateKey) string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return string(privateKeyPEM)
}

// exportRSAPublicKeyAsPEM exports an RSA public key as a PEM string
func exportRSAPublicKeyAsPEM(publicKey *rsa.PublicKey) string {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return string(publicKeyPEM)
}
func OTPEncrypt(message []byte) (key []byte, ciphertext []byte, err error) {
	key = make([]byte, len(message))
	ciphertext = make([]byte, len(message))

	// Generate a random key of the same length as the message
	_, err = io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, nil, err
	}

	// Add each byte of the message with the corresponding byte of the key modulo 256
	for i := range message {
		ciphertext[i] = (message[i] ^ key[i])
	}

	return key, ciphertext, nil
}

func OTPDecrypt(key []byte, ciphertext []byte) (message []byte) {
	message = make([]byte, len(ciphertext))

	// Subtract each byte of the ciphertext with the corresponding byte of the key modulo 256
	for i := range ciphertext {
		message[i] = (ciphertext[i] ^ key[i])
	}

	return message
}
func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	publicKey := &privateKey.PublicKey
	message := fmt.Sprintf("PUBLIC_KEY %s %d\n", publicKey.N.Text(16), publicKey.E)
	content := strings.Split(message, " ")
	modulus := new(big.Int)
	modulus.SetString(content[1], 16) // Base 16 (hexadecimal)
	exponent := new(big.Int)
	exponent.SetString(content[2], 10)
	newPublicKey := &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}

	text := "aku adalah seorang uhuy jagoan"
	key, cipherMessage, err := OTPEncrypt([]byte(text))
	if err != nil {
		panic(err)
	}
	fmt.Printf("otp key: %s\n", key)
	fmt.Printf("ciphered message: %s\n", string(cipherMessage))
	fmt.Println(message)

	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, newPublicKey, []byte(key), nil)
	if err != nil {
		panic(err)
	}
	println(string(encryptedKey))
	encodedEncryptedKey := base64.StdEncoding.EncodeToString(encryptedKey)
	encodedCipherMessage := base64.StdEncoding.EncodeToString(cipherMessage)
	message2 := fmt.Sprintf("MESSAGE %s %s\n", encodedEncryptedKey, encodedCipherMessage)
	parts2 := strings.SplitN(message2, " ", 3)
	decodedEncryptedKey, err := base64.StdEncoding.DecodeString(parts2[1])
	decryptedKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, decodedEncryptedKey, nil)
	if err != nil {
		panic(err)
	}
	println(string(decryptedKey))
	decodedCipherMessage, err := base64.StdEncoding.DecodeString(parts2[2])
	decryptedMessage := OTPDecrypt(decryptedKey, decodedCipherMessage)
	fmt.Printf("decrypted message: %s\n", string(decryptedMessage))
	yey := ""
	fmt.Scan(&yey)
	fmt.Printf("yey: %s\n", yey)
	
}
