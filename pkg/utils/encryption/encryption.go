package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// Interface is a higher level interface for encryption and decryption of arbitrary data
type Interface interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(cipher []byte) ([]byte, error)
}

type encryptionAPI struct {
	gcm cipher.AEAD
}

// NewInterface creates a new symetric encryption interface
func NewInterface(key []byte) (Interface, error) {
	gcm, err := createGCM(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm: %v", err)
	}

	return &encryptionAPI{gcm}, nil
}

func (e *encryptionAPI) Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return e.gcm.Seal(nonce, nonce, data, nil), nil
}

func (e *encryptionAPI) Decrypt(cipher []byte) ([]byte, error) {
	nonceSize := e.gcm.NonceSize()
	if len(cipher) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipher := cipher[:nonceSize], cipher[nonceSize:]
	return e.gcm.Open(nil, nonce, cipher, nil)
}

func createGCM(key []byte) (cipher.AEAD, error) {
	keyLen := len(key)
	switch {
	case keyLen > 32:
		key = key[:32]
	case keyLen > 24:
		key = key[:24]
	case keyLen > 16:
		key = key[:16]
	case keyLen < 16:
		return nil, errors.New("incorrect key: len must be 16, 24 or 32")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}
