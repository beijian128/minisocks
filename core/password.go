// Package core 包包含 minisocks 的核心功能实现
package core

import (
	"encoding/base64" // 提供 Base64 编码和解码功能
	"errors"          // 提供错误处理功能
	"math/rand"       // 提供伪随机数生成功能
	"strings"         // 提供字符串操作功能
	"time"            // 提供时间相关功能
)

// PasswordLength 定义密码的长度，固定为 256 字节
const PasswordLength = 256

// ErrInvalidPassword 表示密码无效的错误
var ErrInvalidPassword = errors.New("invalid password")

// Password 类型定义密码结构，是一个长度为 PasswordLength 的字节数组
type Password [PasswordLength]byte

// init 函数在包被加载时自动执行，用于初始化随机数种子
// 注意：rand.Seed 已弃用，建议使用 crypto/rand 生成安全随机数
func init() {
	rand.Seed(time.Now().Unix())
}

// String 方法将 Password 实例转换为 Base64 编码的字符串
// 返回：Base64 编码后的字符串
func (password *Password) String() string {
	return base64.StdEncoding.EncodeToString(password[:])
}

// ParsePassword 解析 Base64 编码的字符串，将其转换为 Password 实例
// 参数 passwordString：Base64 编码的密码字符串
// 返回：指向 Password 实例的指针和可能出现的错误
func ParsePassword(passwordString string) (*Password, error) {
	// 去除字符串两端的空白字符并进行 Base64 解码
	bs, err := base64.StdEncoding.DecodeString(strings.TrimSpace(passwordString))
	// 若解码出错或解码后的字节切片长度不等于 PasswordLength，则返回错误
	if err != nil || len(bs) != PasswordLength {
		return nil, ErrInvalidPassword
	}
	var password Password
	// 将解码后的字节切片复制到 Password 实例中
	copy(password[:], bs)
	// 释放 bs 引用，帮助垃圾回收
	bs = nil
	return &password, nil
}

// RandPassword 生成一个长度为 256 字节的随机密码，且每个字节位不重复
// 生成的密码会使用 Base64 编码为字符串
// 返回：指向生成的 Password 实例的指针
func RandPassword() *Password {
	// 生成一个包含 0 到 PasswordLength-1 的随机排列的整数切片
	intArr := rand.Perm(PasswordLength)
	password := &Password{}
	// 记录数组索引和对应值相同的元素个数
	sameCount := 0
	for i, v := range intArr {
		password[i] = byte(v)
		if i == v {
			sameCount++
		}
	}
	// 若存在数组索引和对应值相同的元素，则重新生成密码
	if sameCount > 0 {
		password = nil
		return RandPassword()
	} else {
		return password
	}
}
