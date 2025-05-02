package core

type SimpleCi struct {
}

func NewSimple() (*SimpleCi, error) {
	return &SimpleCi{}, nil
}
func (c *SimpleCi) Encrypt(plaintext []byte) ([]byte, error) {
	for i := range plaintext {
		plaintext[i] = plaintext[i] + byte(i)
	}
	return plaintext, nil
}
func (c *SimpleCi) Decrypt(ciphertext []byte) ([]byte, error) {
	for i := range ciphertext {
		ciphertext[i] = ciphertext[i] - byte(i)
	}
	return ciphertext, nil
}
