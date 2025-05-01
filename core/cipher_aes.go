package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

var AESDefaultKey = "37943838a50d61c993e3e6f4f3bd729ff556b973db59852b4c050bb7c6edd699"

// AES 加密器结构体
type AES struct {
	key []byte
}

// NewAES 创建一个新的 AES 加密器实例
// key 必须是 16(AES-128), 24(AES-192) 或 32(AES-256) 字节长度
func NewAES(key []byte) (*AES, error) {
	if len(key) == 0 {
		key, _ = hex.DecodeString(AESDefaultKey)
	}
	switch len(key) {
	case 16, 24, 32:
		return &AES{key: key}, nil
	default:
		return nil, errors.New("invalid key size, must be 16, 24 or 32 bytes")
	}
}

// Encrypt 加密数据
func (a *AES) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt 解密数据
func (a *AES) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateRandomKey 生成随机密钥
// size 必须是 16, 24 或 32
func GenerateRandomKey(size int) ([]byte, error) {
	if size != 16 && size != 24 && size != 32 {
		return nil, errors.New("key size must be 16, 24 or 32 bytes")
	}

	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}
