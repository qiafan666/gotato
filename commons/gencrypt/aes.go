package gencrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gerr"
)

// Md5 returns the md5 hash of the input string.
func Md5(s string, salt ...string) string {
	h := md5.New()
	h.Write([]byte(s))
	if len(salt) > 0 {
		h.Write([]byte(salt[0]))
	}

	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func AesKey(keySize int) (string, error) {
	switch keySize {
	case 16, 24, 32: // 128位、192位和256位密钥长度
		return gcommon.RandStr(keySize), nil
	default:
		return "", gerr.New("GenerateAESKey: invalid key size", "keySize", keySize)
	}
}

func pKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncrypt(origData []byte, key string) (string, error) {
	convertKey := []byte(key)
	block, err := aes.NewCipher(convertKey)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	origData = pKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, convertKey[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func AesDecrypt(crypted string, key string) (text string, err error) {
	defer func() {
		if crash := recover(); crash != nil {
			err = gerr.WrapMsg(err, "AesDecrypt panic", "cryted", crypted, "key", key)
		}
	}()
	convertKey := []byte(key)
	block, err := aes.NewCipher(convertKey)
	if err != nil {
		return
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, convertKey[:blockSize])
	cryptedByte, err := base64.StdEncoding.DecodeString(crypted)
	if err != nil {
		return
	}
	origData := make([]byte, len(cryptedByte))

	blockMode.CryptBlocks(origData, cryptedByte)
	text = string(pKCS7UnPadding(origData))
	return
}
