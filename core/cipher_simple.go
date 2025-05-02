package core

import (
	"encoding/hex"
	"math/rand"
)

type SimpleCi struct {
	a2b [256]byte
	b2a [256]byte
}

func NewSimple(table string) (*SimpleCi, error) {
	a2b, _ := hex.DecodeString(table)
	var b2a [256]byte
	for i := range a2b {
		b2a[a2b[i]] = byte(i)
	}
	return &SimpleCi{
		a2b: [256]byte(a2b),
		b2a: b2a,
	}, nil
}
func (c *SimpleCi) Encrypt(plaintext []byte) ([]byte, error) {
	for i := range plaintext {
		plaintext[i] = c.a2b[plaintext[i]]
	}
	return plaintext, nil
}
func (c *SimpleCi) Decrypt(ciphertext []byte) ([]byte, error) {
	for i := range ciphertext {
		ciphertext[i] = c.b2a[ciphertext[i]]
	}
	return ciphertext, nil
}

func GenerateCipherTable() string {
	var table [256]byte
	r := rand.Perm(256)
	for i := 0; i < 256; i++ {
		table[i] = byte(r[i])
	}
	return hex.EncodeToString(table[:])
}
