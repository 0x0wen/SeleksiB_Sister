package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
)

func GenerateAsymmetricCryptoKeys(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	e := 65537 // Common choice for public exponent
	bigE := big.NewInt(int64(e))
	p, _ := rand.Prime(rand.Reader, bits/2)
	q, _ := rand.Prime(rand.Reader, bits/2)

	n := new(big.Int).Mul(p, q) // defines the max length of the message to be encrypted

	phi := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))

	d := new(big.Int).ModInverse(bigE, phi)
	publicKey := &rsa.PublicKey{
		N: n,
		E: e,
	}

	privateKey := &rsa.PrivateKey{
		PublicKey: *publicKey,
		D:         d,
		Primes:    []*big.Int{p, q},
	}
	return privateKey, publicKey
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

func pkcs7Pad(data []byte, blockSize int) []byte {
    padSize := blockSize - len(data)%blockSize
    pad := bytes.Repeat([]byte{byte(padSize)}, padSize)
    return append(data, pad...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
    if len(data) == 0 || len(data)%blockSize != 0 {
        return nil, fmt.Errorf("invalid padding size")
    }

    padSize := int(data[len(data)-1])
    if padSize > blockSize || padSize == 0 {
        return nil, fmt.Errorf("invalid padding")
    }

    return data[:len(data)-padSize], nil
}

func encryptAES(key, plaintext []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    plaintext = pkcs7Pad(plaintext, aes.BlockSize)
    cipherText := make([]byte, aes.BlockSize+len(plaintext))
    iv := cipherText[:aes.BlockSize]

    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(cipherText[aes.BlockSize:], plaintext)

    return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decryptAES(key []byte, cipherText string) (string, error) {
    encryptedData, err := base64.StdEncoding.DecodeString(cipherText)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    if len(encryptedData) < aes.BlockSize {
        return "", err
    }

    iv := encryptedData[:aes.BlockSize]
    encryptedData = encryptedData[aes.BlockSize:]

    if len(encryptedData)%aes.BlockSize != 0 {
        return "", err
    }

    mode := cipher.NewCBCDecrypter(block, iv)
    mode.CryptBlocks(encryptedData, encryptedData)
	result,err :=pkcs7Unpad(encryptedData, aes.BlockSize)
    return string(result),err
}