package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
)

// xorBytes performs a byte-wise XOR operation on two byte slices.
func xorBytes(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}

// mgf1 is the mask generation function based on SHA-256.
func mgf1(seed []byte, length int) []byte {
	var counter uint32
	mask := make([]byte, length)

	hashLen := sha256.Size
	for i := 0; i < length; i += hashLen {
		counterBytes := []byte{byte(counter >> 24), byte(counter >> 16), byte(counter >> 8), byte(counter)}
		h := sha256.New()
		h.Write(seed)
		h.Write(counterBytes)
		copy(mask[i:], h.Sum(nil))
		counter++
	}

	return mask
}

// oaepPad applies the OAEP padding to the message.
func oaepPad(message, label []byte, k int) ([]byte, error) {
	hash := sha256.New()
	hLen := hash.Size()

	// Check if the message length is valid
	if len(message) > k-2*hLen-2 {
		return nil, fmt.Errorf("message too long")
	}

	lHash := sha256.Sum256(label)
	ps := make([]byte, k-2*hLen-2-len(message))
	db := append(append(lHash[:], ps...), 0x01)
	db = append(db, message...)

	seed := make([]byte, hLen)
	_, err := rand.Read(seed)
	if err != nil {
		return nil, err
	}

	dbMask := mgf1(seed, len(db))
	maskedDB := xorBytes(db, dbMask)
	seedMask := mgf1(maskedDB, hLen)
	maskedSeed := xorBytes(seed, seedMask)

	em := append(append([]byte{0x00}, maskedSeed...), maskedDB...)
	return em, nil
}

// rsaEncrypt performs the RSA encryption: ciphertext = (message^e) mod n
func rsaEncrypt(message []byte, e int, n *big.Int) *big.Int {
	m := new(big.Int).SetBytes(message)
	c := new(big.Int).Exp(m, big.NewInt(int64(e)), n)
	return c
}

// implements the RSA-OAEP encryption.
func EncryptOAEP(message, label []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	k := (pubKey.N.BitLen() + 7) / 8

	// Apply OAEP padding
	em, err := oaepPad(message, label, k)
	if err != nil {
		return nil, err
	}

	// Perform RSA encryption
	ciphertext := rsaEncrypt(em, pubKey.E, pubKey.N)
	return ciphertext.Bytes(), nil
}
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

// oaepUnpad reverses the OAEP padding process and extracts the original message.
func oaepUnpad(em, label []byte, k int) ([]byte, error) {
	hash := sha256.New()
	hLen := hash.Size()

	if len(em) != k {
		return nil, fmt.Errorf("decryption error")
	}

	lHash := sha256.Sum256(label)

	maskedSeed := em[1 : 1+hLen]
	maskedDB := em[1+hLen:]

	seedMask := mgf1(maskedDB, hLen)
	seed := xorBytes(maskedSeed, seedMask)

	dbMask := mgf1(seed, len(maskedDB))
	db := xorBytes(maskedDB, dbMask)

	// Validate the hash of the label
	if !equal(db[:hLen], lHash[:]) {
		return nil, fmt.Errorf("decryption error")
	}

	// Remove the padding
	for i := hLen; i < len(db); i++ {
		if db[i] == 0x01 {
			return db[i+1:], nil
		}
	}

	return nil, fmt.Errorf("decryption error")
}

// rsaDecrypt performs the RSA decryption: message = (ciphertext^d) mod n
func rsaDecrypt(ciphertext *big.Int, d, n *big.Int) *big.Int {
	m := new(big.Int).Exp(ciphertext, d, n)
	return m
}

// DecryptOAEP manually implements the RSA-OAEP decryption.
func DecryptOAEP(ciphertext []byte, label []byte, privKey *rsa.PrivateKey) ([]byte, error) {
	k := (privKey.N.BitLen() + 7) / 8

	// Convert ciphertext to a big integer
	c := new(big.Int).SetBytes(ciphertext)

	// Perform RSA decryption
	m := rsaDecrypt(c, privKey.D, privKey.N)

	// Convert the decrypted message to bytes
	em := m.Bytes()

	// Ensure the decrypted message is the correct length
	if len(em) < k {
		paddedEm := make([]byte, k)
		copy(paddedEm[k-len(em):], em)
		em = paddedEm
	}

	// Unpad the message using OAEP
	message, err := oaepUnpad(em, label, k)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Helper function to check if two byte slices are equal
func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}