package ports

type Encryptor interface {
	Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error)
	Decrypt(ciphertext, nonce []byte) ([]byte, error)
	KeyVersion() int
}
