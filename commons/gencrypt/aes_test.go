package gencrypt

import (
	"testing"
)

const testText string = "{\"userId\":1,\"amount\"}"

func TestAesEncrypt(t *testing.T) {

	key, err := AesKey(32)
	if err != nil {
		return
	}
	t.Log("AES Key = ", key)

	encrypt, err := AesEncrypt([]byte(testText), key)
	if err != nil {
		return
	}
	t.Log("AES Encrypt = ", encrypt)

	decrypt, err := AesDecrypt(encrypt, key)
	if err != nil {
		return
	}
	t.Log("AES Decrypt = ", decrypt)
}

func TestGenerateAESKey(t *testing.T) {
	key, err := AesKey(16)
	if err != nil {
		t.Log("GenerateAESKey() error = ", err)
	}
	t.Log("AES 16 Key = ", key)

	key, err = AesKey(24)
	if err != nil {
		t.Log("GenerateAESKey() error = ", err)
	}
	t.Log("AES 24 Key = ", key)

	key, err = AesKey(32)
	if err != nil {
		t.Log("GenerateAESKey() error = ", err)
	}
	t.Log("AES 32 Key = ", key)
}
