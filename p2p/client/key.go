package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

func storeRSAKeys(privateKey *rsa.PrivateKey) {
    privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
    publicKeyBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	err := storeKey(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}), "private_key.txt")
	handleErr(err)
	err = storeKey(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: publicKeyBytes}), "public_key.txt")
	handleErr(err)
}

func storeAESKey() ([]byte, error) {
    key := make([]byte, 32) // AES-256 key size
    _, err := rand.Read(key)
    if err != nil {
        return nil, err
    }
	err = storeKey(key, "aes_key.txt")
	handleErr(err)
    return key, nil
}

func storeKey(key []byte, filePath string) error {
    return os.WriteFile(filePath, key, 0600)
}


func loadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func loadPublicKey(filePath string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	return x509.ParsePKCS1PublicKey(block.Bytes)
}

func loadAESKey(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func getOrCreateRSAKey(filePath string) (*rsa.PrivateKey, error) {
    // Check if the file exists
    if _, err := os.Stat(filePath); err == nil {
        // File exists, load the private key
        return loadPrivateKey(filePath)
    }

    // File doesn't exist, generate a new private key
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return nil, err
    }

    // Save the private key to the file
   	storeRSAKeys(privateKey)
    if err != nil {
        return nil, err
    }

    return privateKey, nil
}

func getOrCreateAESKey(filePath string) ([]byte, error) {
    // Check if the file exists
    if _, err := os.Stat(filePath); err == nil {
        // File exists, load the AES key
        return loadAESKey(filePath)
    }

    // File doesn't exist, generate a new AES key
    aesKey := make([]byte, 32) // 256-bit key
    _, err := rand.Read(aesKey)
    if err != nil {
        return nil, err
    }

    // Save the AES key to the file
    err = os.WriteFile(filePath, aesKey, 0600)
    if err != nil {
        return nil, err
    }

    return aesKey, nil
}