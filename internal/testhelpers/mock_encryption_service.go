package testhelpers

type mockEncryptionService struct {
}

func NewMockPassthroughEncryptionService() *mockEncryptionService {
	return &mockEncryptionService{}
}

func (s *mockEncryptionService) EncryptBinary(plaintext []byte) ([]byte, string, error) {
	return plaintext, "mockKeyId", nil
}

func (s *mockEncryptionService) DecryptBinary(ciphertext []byte, keyId string) ([]byte, error) {
	return ciphertext, nil
}

func (s *mockEncryptionService) Encrypt(plaintext string) (string, string, error) {
	return plaintext, "mockKeyId", nil
}

func (s *mockEncryptionService) Decrypt(ciphertext string, keyId string) (string, error) {
	return ciphertext, nil
}
