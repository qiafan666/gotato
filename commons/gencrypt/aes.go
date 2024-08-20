package gencrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
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

func GenerateAESKey(keySize int) (string, error) {
	switch keySize {
	case 16, 24, 32: // 128位、192位和256位密钥长度
		key := make([]byte, keySize)
		_, err := rand.Read(key)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(key), nil
	default:
		return "", errors.New("unsupported key size")
	}
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncrypt(origData, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func AesDecrypt(crypted string, key []byte) (text string, err error) {
	defer func() {
		if crash := recover(); crash != nil {
			err = errors.New("crypted is aes text")
		}
	}()
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	cryptedByte, err := base64.StdEncoding.DecodeString(crypted)
	if err != nil {
		return
	}
	origData := make([]byte, len(cryptedByte))

	blockMode.CryptBlocks(origData, cryptedByte)
	text = string(PKCS7UnPadding(origData))
	return
}
