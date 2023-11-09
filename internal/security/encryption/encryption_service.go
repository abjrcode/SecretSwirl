package encryption

type EncryptionService interface {
	EncryptBinary(plaintext []byte) ([]byte, string, error)
	DecryptBinary(ciphertext []byte, keyId string) ([]byte, error)
	Encrypt(plaintext string) (string, string, error)
	Decrypt(ciphertext, keyId string) (string, error)
}
