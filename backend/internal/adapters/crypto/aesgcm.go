package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

const currentKeyVersion = 1

var ErrBadCiphertext = errors.New("bad ciphertext")

type AESGCM struct {
	aead cipher.AEAD
}

func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AESGCM{aead: aead}, nil
}

func (a *AESGCM) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	nonce = make([]byte, a.aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	ciphertext = a.aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func (a *AESGCM) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
	if len(nonce) != a.aead.NonceSize() {
		return nil, ErrBadCiphertext
	}
	plaintext, err := a.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrBadCiphertext
	}
	return plaintext, nil
}

func (a *AESGCM) KeyVersion() int { return currentKeyVersion }
