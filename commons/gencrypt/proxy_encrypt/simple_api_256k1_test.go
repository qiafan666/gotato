package proxy_encrypt

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAPIBasics(t *testing.T) {

	cxt := MakeDefaultContext()

	//Alice
	privKeyAlice := GenPrivateKey(cxt)
	pubKeyAlice := privKeyAlice.GetPublicKey(cxt)

	str := StructToString(pubKeyAlice)
	fmt.Println(str, len(str))

	fmt.Println(len(priPrefix), len(priPostfix), len(pubPreFix), len(pubPostFix))

	fmt.Println(privKeyAlice.String(), len(privKeyAlice.String()))

	fmt.Println(pubKeyAlice.String(), len(pubKeyAlice.String()))

	//a=StructToString(pubKeyAlice2))
	//fmt.Println(len(StructToString(privKeyAlice)))
	var obj = new(UmbralCurveElement)
	StringToStruct(StructToString(pubKeyAlice), obj)
	fmt.Println(obj.String())
	fmt.Println(pubKeyAlice.String())
	//fmt.Println(obj.GetPublicKey(cxt))
	//fmt.Println(cxt.curveField.GetGen())
	//fmt.Println(pubKeyAlice.MulScalar(privKeyAlice.Invert().GetValue()))

	msg := []byte("hellohellohellohellohellohellohello")
	r, s := Sign(cxt, msg, privKeyAlice)
	fmt.Println(Verify(cxt, r, s, msg, pubKeyAlice))

	//Bob
	privKeyBob := GenPrivateKey(cxt)
	pubKeyBob := privKeyBob.GetPublicKey(cxt)

	plainText := []byte("attack at dawn")
	//Charlie or Alice
	//胶囊中包含在解密期间重新生成 新密钥的必要信息
	cipherText, capsule := Encrypt(cxt, pubKeyAlice, plainText)

	//Alice

	testDecrypt := DecryptDirect(cxt, capsule, privKeyAlice, cipherText)

	if !reflect.DeepEqual(plainText, testDecrypt) {
		t.Errorf("Direct decryption failed")
	}
	n := 20
	th := 10
	//Alice
	//kFrags := ReKeyGen(cxt, privKeyAlice, pubKeyBob, 10, 20)
	//
	//Proxy
	privaProxyList := make([]*UmbralFieldElement, n)
	pubProxyList := make([]*UmbralCurveElement, n)
	for i := 0; i < n; i++ {
		privaProxyList[i] = GenPrivateKey(cxt)
		pubProxyList[i] = privaProxyList[i].GetPublicKey(cxt)
	}
	//Alice
	ckFrags := GenerateCKFrags(cxt, privKeyAlice, pubKeyBob, pubProxyList, th, n)

	ckFragStr := StructToString(ckFrags)
	fmt.Println("ckFrags:", ckFrags[0])

	var ckFrags2 = make([]*CKFrag, 1)
	StringToStruct(ckFragStr, &ckFrags2)
	fmt.Println("ckFrags2:", ckFrags2[0])

	/*
		fmt.Println(ckFrags[0])
		data := CKFragToBytes(ckFrags[0])
		fmt.Println(data)
		ckfrag := BytesToCKFrag(data)
		fmt.Println(ckfrag)
	*/

	//Alice sends cFrags[i] to Proxy[i]
	//cFrags[i]

	//Proxy
	kFrags := GetKFrags(cxt, privaProxyList, ckFrags)
	cFrags := make([]*CFrag, th)
	for i := range cFrags {
		cFrags[i] = ReEncapsulate(kFrags[i], capsule)
	}

	//Proxy[i] sends cFrags[i] to Bob
	//Bob
	testDecryptFrags := DecryptFragments(cxt, capsule, cFrags, privKeyBob, pubKeyAlice, cipherText)
	if !reflect.DeepEqual(plainText, testDecryptFrags) {
		t.Errorf("Re-encapsulated fragment decryption failed")
	}
}
