package gcommon

import (
	"bytes"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	mathRand "math/rand"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

// StringsToString 字符串数组转字符串
func StringsToString(stringArray []string) string {
	if len(stringArray) <= 0 {
		return ""
	}
	return strings.Join(stringArray, ",")
}

// StringToStrings 字符串转字符串数组;,逗号分隔
func StringToStrings(param string) []string {
	if len(param) <= 0 {
		return []string{}
	}
	return strings.Split(param, ",")
}

// StringToSha256 字符串转SHA256
func StringToSha256(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte("hello world\n")))
}

// Nonce 生成随机串(size应为偶数)
func Nonce(size uint8) string {
	nonce := make([]byte, size/2)
	io.ReadFull(cryptoRand.Reader, nonce)
	return hex.EncodeToString(nonce)
}

const randomLowerUpperNumberMixed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const randomLower = "abcdefghijklmnopqrstuvwxyz"
const randomUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const randomNumber = "0123456789"
const randomLowerUpper = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandomLowerUpperNumberMixed 生成随机字符串
func RandomLowerUpperNumberMixed(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for j := range b {
			b[j] = randomLowerUpperNumberMixed[mathRand.Intn(len(randomLowerUpperNumberMixed))]
		}
	}
	return string(b)
}

// RandomLower 生成随机小写
func RandomLower(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for j := range b {
			b[j] = randomLower[mathRand.Intn(len(randomLower))]
		}
	}
	return string(b)
}

// RandomUpper 生成随机大写字母
func RandomUpper(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for j := range b {
			b[j] = randomUpper[mathRand.Intn(len(randomUpper))]
		}
	}
	return string(b)
}

// RandomNumber 生成随机数字
func RandomNumber(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for j := range b {
			b[j] = randomNumber[mathRand.Intn(len(randomNumber))]
		}
	}
	return string(b)
}

// RandomLowerUpperMixed 生成大小写英文混合随机字符串
func RandomLowerUpperMixed(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for j := range b {
			b[j] = randomLowerUpper[mathRand.Intn(len(randomLowerUpper))]
		}
	}
	return string(b)
}

// DataCheck 检查输入的字符串是否有空值
func DataCheck(input ...string) []int {
	var nullIndices []int

	for i, data := range input {
		if len(data) == 0 {
			nullIndices = append(nullIndices, i)
		}
	}

	return nullIndices
}

// StringToBytes 原地转换
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2BytesNoCopy
// 无拷贝 string 转 []byte
func String2BytesNoCopy(s string) []byte {
	tmp := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{tmp[0], tmp[1], tmp[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// UnderscoreName 驼峰式写法转为下划线写法
func UnderscoreName(name string) string {
	buffer := NewBuffer()
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.Append('_')
			}
			buffer.Append(unicode.ToLower(r))
		} else {
			buffer.Append(r)
		}
	}

	return buffer.String()
}

// Byte2Uint32 字节数组转uint32
func Byte2Uint32(b []byte) uint32 {
	seed := uint32(131)
	hash := uint32(0)

	for _, v := range b {
		hash = hash*seed + uint32(v)
	}
	return hash
}

// String2Uint32 string 转 uint32
func String2Uint32(s string) uint32 {
	b := bytes.NewBufferString(s).Bytes()
	return Byte2Uint32(b)
}

// CountRune 计算包含中文的字符串长度，但是一个中文算2个长度
func CountRune(str string) int {
	length := 0
	for _, runeValue := range str {
		if utf8.RuneLen(runeValue) > 1 {
			length += 2
		} else {
			length += 1
		}
	}
	return length
}

// ParseChinese 提取字符串中的中文
func ParseChinese(str string) string {
	b := strings.Builder{}
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			b.WriteRune(v)
		}
	}
	return b.String()
}
