package core

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateRandomKey(t *testing.T) {
	got, err := GenerateRandomKey(32)
	assert.NoError(t, err)
	t.Logf("got:%0x, err:%v", got, err)
	t.Logf("got:%v, err:%v", got, err)
	t.Logf("got:%s, err:%v", hex.EncodeToString(got), err)
	t.Log(hex.DecodeString(hex.EncodeToString(got)))

}

func TestAES_Encrypt(t *testing.T) {
	aes, err := NewAES([]byte("12345678901234567890123456789012"))
	assert.NoError(t, err)
	got, err := aes.Encrypt([]byte("hello world"))
	assert.NoError(t, err)

	got2, err := aes.Decrypt(got)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(got2))
}
