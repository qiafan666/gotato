package gcommon

import (
	"crypto/md5"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	mathRand "math/rand"
	"strings"
	"time"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

// StringToMd5 字符串转MD5
func StringToMd5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
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
		for i := range b {
			b[i] = randomLowerUpperNumberMixed[mathRand.Intn(len(randomLowerUpperNumberMixed))]
		}
	}
	return string(b)
}

// RandomLower 生成随机小写
func RandomLower(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomLower[mathRand.Intn(len(randomLower))]
		}
	}
	return string(b)
}

// RandomUpper 生成随机大写字母
func RandomUpper(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomUpper[mathRand.Intn(len(randomUpper))]
		}
	}
	return string(b)
}

// RandomNumber 生成随机数字
func RandomNumber(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomNumber[mathRand.Intn(len(randomNumber))]
		}
	}
	return string(b)
}

// RandomLowerUpperMixed 生成大小写英文混合随机字符串
func RandomLowerUpperMixed(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomLowerUpper[mathRand.Intn(len(randomLowerUpper))]
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
