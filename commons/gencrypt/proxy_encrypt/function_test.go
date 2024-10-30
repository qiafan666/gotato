package proxy_encrypt

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFunction256k1(t *testing.T) {
	cxt := MakeDefaultContext()

	//Alice
	alice := NewProxyEncrypt(Encrypt256k1)

	fmt.Println(len(alice.PrivateKeyString))
	fmt.Println(len(alice.PublicKeyString))

	alice.PrivateKey = alice.String2Pri(alice.Pri2String())

	msg := []byte("hellohellohellohellohellohellohello")
	r, s := alice.Sign(msg)
	fmt.Println(alice.Verify(r, s, msg))

	//Bob
	bob := NewProxyEncrypt(Encrypt256k1)

	plainText := []byte("attack at dawn")
	//Charlie or Alice
	//胶囊中包含在解密期间重新生成 新密钥的必要信息
	cipherText, capsule, _ := alice.Encrypt(plainText)
	//Alice
	decrypt, _ := alice.Decrypt(capsule, cipherText)

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
	ckFrags := GenerateCKFrags(cxt, alice.PrivateKey, bob.PublicKey, pubProxyList, th, n)

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

	fragments, _ := bob.DecryptFragments(capsule, cFrags, alice.PublicKey, cipherText)
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
	cipherText, capsule, _ := encryptAlice.Encrypt(plainText)

	//Alice
	decrypt, _ := encryptAlice.Decrypt(capsule, cipherText)
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
	fragments, _ := encryptBob.DecryptFragments(capsule, cFrags, encryptAlice.PublicKey, cipherText)
	if !reflect.DeepEqual(plainText, fragments) {
		t.Errorf("Re-encapsulated fragment decryption failed")
	}
}
func TestFunction2SM2(t *testing.T) {
	cxt := MakeSM2Context()

	//Alice
	alice := NewProxyEncrypt()

	fmt.Println(len(alice.PrivateKeyString))
	fmt.Println(len(alice.PublicKeyString))

	alice.PrivateKey = alice.String2Pri(alice.Pri2String())
	alice.PublicKey = alice.String2Pub(alice.Pub2String())

	msg := []byte("hellohellohellohellohellohellohello")
	r, s := alice.Sign(msg)
	fmt.Println(alice.Verify(r, s, msg))

	//Bob
	bob := NewProxyEncrypt()

	plainText := []byte("attack at dawn")
	//Charlie or Alice
	//胶囊中包含在解密期间重新生成 新密钥的必要信息
	cipherText, capsule, _ := alice.Encrypt(plainText)

	//Alice
	decrypt, _ := alice.Decrypt(capsule, cipherText)
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
	ckFrags := GenerateCKFrags(cxt, alice.PrivateKey, bob.PublicKey, pubProxyList, th, n)

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
	fragments, _ := bob.DecryptFragments(capsule, cFrags, alice.PublicKey, cipherText)
	if !reflect.DeepEqual(plainText, fragments) {
		t.Errorf("Re-encapsulated fragment decryption failed")
	}
}
