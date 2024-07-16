package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestRsa(t *testing.T) {

	priv, pub, err := GenRsaKey(RSA_1, PKCS_8)
	if err != nil {
		return
	}
	fmt.Println("生成的私钥=" + hex.EncodeToString(priv))
	fmt.Println("生成的公钥=" + hex.EncodeToString(pub))

	data := bytes.Buffer{}
	data.WriteString("1")
	data.WriteString("2")
	data.WriteString("3")
	data.WriteString("4")
	data.WriteString("5")

	sign, err := Rsa2Sign(data.Bytes(), priv, PKCS_8)
	if err != nil {
		fmt.Println(err)
	}
	signData := hex.EncodeToString(sign)
	fmt.Println("签名后的数据=" + signData)

	//decodeString, err := base64.StdEncoding.DecodeString(signData)
	decodeString, err := hex.DecodeString(signData)
	if err != nil {
		fmt.Println(err)
	}
	err = Rsa2VerifySign(sha256.Sum256(data.Bytes()), decodeString, pub)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("test suc..........")
}
