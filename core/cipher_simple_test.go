package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleCi_Encrypt(t *testing.T) {
	var datas [][]byte
	{
		aes, err := NewSimple()
		assert.NoError(t, err)
		for i := 0; i < 10; i++ {
			got, err := aes.Encrypt([]byte(fmt.Sprintf("hello world %d", i)))
			assert.NoError(t, err)
			datas = append(datas, got)
		}

	}

	{

		aes, err := NewSimple()
		assert.NoError(t, err)
		for i := 9; i >= 0; i-- {
			got, err := aes.Decrypt(datas[i])
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("hello world %d", i), string(got))
		}

	}
}
