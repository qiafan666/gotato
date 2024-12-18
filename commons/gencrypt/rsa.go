package gencrypt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"github.com/qiafan666/gotato/commons/gerr"
)

const (
	PKCS_1 = iota + 1
	PKCS_8
)
const (
	RSA_1 = iota + 1
	RSA_2
)

// GenRsaKey 生成RSA密钥
func GenRsaKey(rsaType, keyType int) (prvKey, pubKey []byte, err error) {
	rsaLen := 2048
	if rsaType == RSA_1 {
		rsaLen = 1024
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaLen)
	if err != nil {
		return
	}

	var derStream []byte
	if keyType == PKCS_1 {
		derStream = x509.MarshalPKCS1PrivateKey(privateKey)
	} else if keyType == PKCS_8 {
		derStream, _ = x509.MarshalPKCS8PrivateKey(privateKey)
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvKey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)

	if err != nil {
		return
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubKey = pem.EncodeToMemory(block)
	return
}

// Rsa2Sign 签名
func Rsa2Sign(data []byte, keyBytes []byte, keyType int) (signature []byte, err error) {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return
	}
	var privateKey *rsa.PrivateKey
	if keyType == PKCS_1 {
		var err error
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return []byte(""), err
		}
	} else if keyType == PKCS_8 {
		var err error
		tempKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return []byte(""), err
		}
		privateKey = tempKey.(*rsa.PrivateKey)
	}

	signature, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return
	}

	return
}

// Rsa2VerifySign 验证签名
func Rsa2VerifySign(data [sha256.Size]byte, signData, keyBytes []byte) error {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return gerr.New("private key error!")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return gerr.WrapMsg(err, "public key error!")
	}

	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, data[:], signData)
	if err != nil {
		return gerr.WrapMsg(err, "sign verify error!")
	}
	return nil
}

// RsaEncrypt 加密
func RsaEncrypt(data, keyBytes []byte) (cipherText []byte, err error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, gerr.New("public key error!")
	}
	cipherText, err = rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		return
	}
	return cipherText, err
}

// RsaDecrypt 解密
func RsaDecrypt(cipherText, keyBytes []byte, keyType int) (date []byte, err error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, gerr.New("private key error!")
	}
	var privateKey *rsa.PrivateKey

	if keyType == PKCS_1 {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)

	} else if keyType == PKCS_8 {
		var tempKey interface{}
		tempKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		privateKey = tempKey.(*rsa.PrivateKey)
	}
	if err != nil {
		return nil, err
	}
	// Decrypt
	data, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	if err != nil {
		return nil, err
	}
	return data, err
}
