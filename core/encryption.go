package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

// Generate a 32-byte key from a password using SHA-256
func generateKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	// fmt.Printf("Generated Key: %x\n", hash[:]) // Log the generated key
	return hash[:]
}

// Encrypt data using AES-256
func encrypt(data []byte, password string) (string, error) {
	key := generateKey(password) // Use password to generate key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encryptedData := gcm.Seal(nonce, nonce, data, nil)
	encryptedStr := base64.URLEncoding.EncodeToString(encryptedData)
	// fmt.Printf("Encrypted Data: %s\n", encryptedStr)
	return encryptedStr, nil
}

// Decrypt data using AES-256
func decrypt(data string, password string) ([]byte, error) {
	key := generateKey(password) // Use password to generate key
	// fmt.Printf("Key for Decryption: %x\n", key)
	encryptedData, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	decryptedData, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
