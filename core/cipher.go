package core

// Password 表示密码类型，这里假设 Password 是一个字节切片类型，
// 实际代码中可能需要在别处定义该类型，例如 type Password [256]byte
// 这里注释说明，方便理解代码逻辑
// type Password [256]byte

// Cipher 结构体表示一个编解码器，包含编码和解码所需的密码。
type Cipher struct {
	// encodePassword 是用于编码原始数据的密码。
	encodePassword *Password
	// decodePassword 是用于解码加密后数据的密码。
	decodePassword *Password
}

// encode 方法对传入的字节切片进行编码操作。
// 参数 bs 是需要编码的原始数据字节切片。
func (cipher *Cipher) encode(bs []byte) {
	// 遍历字节切片中的每个字节
	for i, v := range bs {
		// 根据编码密码替换当前字节的值
		bs[i] = cipher.encodePassword[v]
	}
}

// decode 方法对传入的字节切片进行解码操作，将加密后的数据还原为原始数据。
// 参数 bs 是需要解码的加密数据字节切片。
func (cipher *Cipher) decode(bs []byte) {
	// 遍历字节切片中的每个字节
	for i, v := range bs {
		// 根据解码密码替换当前字节的值
		bs[i] = cipher.decodePassword[v]
	}
}

// NewCipher 创建一个新的编解码器实例。
// 参数 encodePassword 是用于编码的密码指针。
// 返回一个指向新创建的 Cipher 结构体的指针。
func NewCipher(encodePassword *Password) *Cipher {
	// 初始化解码密码
	decodePassword := &Password{}
	// 遍历编码密码，构建解码密码映射
	for i, v := range encodePassword {
		// 这行代码实际是冗余的，可删除，encodePassword[i] 本身就是 v
		// encodePassword[i] = v
		decodePassword[v] = byte(i)
	}
	return &Cipher{
		encodePassword: encodePassword,
		decodePassword: decodePassword,
	}
}
