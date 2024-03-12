package CSCrypto

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFunction256k1(t *testing.T) {
	cxt := MakeDefaultContext()

	//Alice
	encryptAlice := NewProxyEncrypt(Encrypt256k1)

	fmt.Println(len(encryptAlice.PrivateKeyString))
	fmt.Println(len(encryptAlice.PublicKeyString))

	msg := []byte("hellohellohellohellohellohellohello")
	r, s := Sign(cxt, msg, encryptAlice.PrivateKey)
	fmt.Println(Verify(cxt, r, s, msg, encryptAlice.PublicKey))

	//Bob
	encryptBob := NewProxyEncrypt(Encrypt256k1)

	plainText := []byte("attack at dawn")
	//Charlie or Alice
	//胶囊中包含在解密期间重新生成 新密钥的必要信息
	cipherText, capsule := encryptAlice.Encrypt(plainText)
	//Alice
	decrypt := encryptAlice.Decrypt(capsule, cipherText)

	if !reflect.DeepEqual(plainText, decrypt) {
		t.Errorf("Direct decryption failed")
	}

	n := 20
	th := 10

	//GenProxy
	privaProxyList := make([]*UmbralFieldElement, n)
	pubProxyList := make([]*UmbralCurveElement, n)
	for i := 0; i < n; i++ {
		encrypt := NewProxyEncrypt(Encrypt256k1)
		privaProxyList[i] = encrypt.PrivateKey
		pubProxyList[i] = encrypt.PublicKey
	}
	//Alice
	ckFrags := GenerateCKFrags(cxt, encryptAlice.PrivateKey, encryptBob.PublicKey, pubProxyList, th, n)

	ckFragStr := StructToString(ckFrags)
	fmt.Println("ckFrags:", ckFrags[0])

	var ckFrags2 = make([]*CKFrag, 1)
	StringToStruct(ckFragStr, &ckFrags2)
	fmt.Println("ckFrags2:", ckFrags2[0])

	//Proxy
	kFrags := GetKFrags(cxt, privaProxyList, ckFrags)
	cFrags := make([]*CFrag, th)
	for i := range cFrags {
		cFrags[i] = ReEncapsulate(kFrags[i], capsule)
	}

	//Proxy[i] sends cFrags[i] to Bob
	//Bob

	fragments := encryptBob.DecryptFragments(capsule, cFrags, encryptAlice.PublicKey, cipherText)
	if !reflect.DeepEqual(plainText, fragments) {
		t.Errorf("Re-encapsulated fragment decryption failed")
	}
}
func TestFunctionSM2(t *testing.T) {
	cxt := MakeSM2Context()

	//Alice
	encryptAlice := NewProxyEncrypt(EncryptSm2)

	fmt.Println(len(encryptAlice.PrivateKeyString))
	fmt.Println(len(encryptAlice.PublicKeyString))

	msg := []byte("hellohellohellohellohellohellohello")
	r, s := Sign(cxt, msg, encryptAlice.PrivateKey)
	fmt.Println(Verify(cxt, r, s, msg, encryptAlice.PublicKey))

	//Bob
	encryptBob := NewProxyEncrypt(EncryptSm2)

	plainText := []byte("attack at dawn")
	//Charlie or Alice
	//胶囊中包含在解密期间重新生成 新密钥的必要信息
	cipherText, capsule := encryptAlice.Encrypt(plainText)

	//Alice
	decrypt := encryptAlice.Decrypt(capsule, cipherText)
	if !reflect.DeepEqual(plainText, decrypt) {
		t.Errorf("Direct decryption failed")
	}

	n := 20
	th := 10

	//GenProxy
	privaProxyList := make([]*UmbralFieldElement, n)
	pubProxyList := make([]*UmbralCurveElement, n)
	for i := 0; i < n; i++ {
		encrypt := NewProxyEncrypt(EncryptSm2)
		privaProxyList[i] = encrypt.PrivateKey
		pubProxyList[i] = encrypt.PublicKey
	}
	//Alice
	ckFrags := GenerateCKFrags(cxt, encryptAlice.PrivateKey, encryptBob.PublicKey, pubProxyList, th, n)

	ckFragStr := StructToString(ckFrags)
	fmt.Println("ckFrags:", ckFrags[0])

	var ckFrags2 = make([]*CKFrag, 1)
	StringToStruct(ckFragStr, &ckFrags2)
	fmt.Println("ckFrags2:", ckFrags2[0])

	//Proxy
	kFrags := GetKFrags(cxt, privaProxyList, ckFrags)
	cFrags := make([]*CFrag, th)
	for i := range cFrags {
		cFrags[i] = ReEncapsulate(kFrags[i], capsule)
	}

	//Proxy[i] sends cFrags[i] to Bob
	//Bob
	fragments := encryptBob.DecryptFragments(capsule, cFrags, encryptAlice.PublicKey, cipherText)
	if !reflect.DeepEqual(plainText, fragments) {
		t.Errorf("Re-encapsulated fragment decryption failed")
	}
}
func TestFunction2SM2(t *testing.T) {
	cxt := MakeSM2Context()

	//Alice
	encryptAlice := NewProxyEncrypt()

	fmt.Println(len(encryptAlice.PrivateKeyString))
	fmt.Println(len(encryptAlice.PublicKeyString))

	msg := []byte("hellohellohellohellohellohellohello")
	r, s := Sign(cxt, msg, encryptAlice.PrivateKey)
	fmt.Println(Verify(cxt, r, s, msg, encryptAlice.PublicKey))

	//Bob
	encryptBob := NewProxyEncrypt()

	plainText := []byte("attack at dawn")
	//Charlie or Alice
	//胶囊中包含在解密期间重新生成 新密钥的必要信息
	cipherText, capsule := encryptAlice.Encrypt(plainText)

	//Alice
	decrypt := encryptAlice.Decrypt(capsule, cipherText)
	if !reflect.DeepEqual(plainText, decrypt) {
		t.Errorf("Direct decryption failed")
	}

	n := 20
	th := 10

	//GenProxy
	privaProxyList := make([]*UmbralFieldElement, n)
	pubProxyList := make([]*UmbralCurveElement, n)
	for i := 0; i < n; i++ {
		encrypt := NewProxyEncrypt(EncryptSm2)
		privaProxyList[i] = encrypt.PrivateKey
		pubProxyList[i] = encrypt.PublicKey
	}
	//Alice
	ckFrags := GenerateCKFrags(cxt, encryptAlice.PrivateKey, encryptBob.PublicKey, pubProxyList, th, n)

	ckFragStr := StructToString(ckFrags)
	fmt.Println("ckFrags:", ckFrags[0])

	var ckFrags2 = make([]*CKFrag, 1)
	StringToStruct(ckFragStr, &ckFrags2)
	fmt.Println("ckFrags2:", ckFrags2[0])

	//Proxy
	kFrags := GetKFrags(cxt, privaProxyList, ckFrags)
	cFrags := make([]*CFrag, th)
	for i := range cFrags {
		cFrags[i] = ReEncapsulate(kFrags[i], capsule)
	}

	//Proxy[i] sends cFrags[i] to Bob
	//Bob
	fragments := encryptBob.DecryptFragments(capsule, cFrags, encryptAlice.PublicKey, cipherText)
	if !reflect.DeepEqual(plainText, fragments) {
		t.Errorf("Re-encapsulated fragment decryption failed")
	}
}
