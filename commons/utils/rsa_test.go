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
	data := fmt.Sprintf("%s%s%s%s%s", "1", "2", "3", "4", "5")

	dataString := bytes.NewBufferString(data)
	sign, err := Rsa2Sign(dataString.Bytes(), priv, PKCS_8)
	if err != nil {
		fmt.Println(err)
	}
	encryptOut := hex.EncodeToString(sign)
	fmt.Println("加密后的数据=" + encryptOut)

	bufferString := bytes.Buffer{}
	bufferString.WriteString("1")
	bufferString.WriteString("2")
	bufferString.WriteString("3")
	bufferString.WriteString("4")
	bufferString.WriteString("5")

	//decodeString, err := base64.StdEncoding.DecodeString(encryptOut)
	decodeString, err := hex.DecodeString(encryptOut)
	if err != nil {
		fmt.Println(err)
	}
	err = Rsa2VerifySign(sha256.Sum256(bufferString.Bytes()), decodeString, pub)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("test suc..........")
}
