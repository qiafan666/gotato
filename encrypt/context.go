package CSCrypto

import (
	"fmt"
	"math/big"
	"reflect"
)

type Context struct {
	curveField   *CurveField
	targetField  *ZField
	minValSha512 *big.Int
	U            *CurveElement
	symKeySize   int
}

func (cxt *Context) GetGen() *CurveElement {
	return cxt.curveField.GetGen()
}

func (cxt *Context) GetOrder() *big.Int {
	return cxt.curveField.FieldOrder
}

func (cxt *Context) MulGen(x *ModInt) *UmbralCurveElement {
	return &UmbralCurveElement{*cxt.GetGen().MulScalar(x.GetValue())}
}

func (cxt *Context) MulU(x *ModInt) *UmbralCurveElement {
	return &UmbralCurveElement{*cxt.U.MulScalar(x.GetValue())}
}

func getMinValSha512(curve *CurveField) *big.Int {
	maxInt512 := big.NewInt(0).Lsh(big.NewInt(1), 512)
	return big.NewInt(0).Mod(maxInt512, curve.FieldOrder)
}

const SECRET_BOX_KEY_SIZE = 32
const SECRET_SM4_KEY_SIZE = 16

func MakeDefaultContext() *Context {
	curveField := MakeSecp256k1()
	targetField := MakeZField(curveField.FieldOrder)
	uX, _ := big.NewInt(0).SetString("68282748765985831108782504644936740559294230795844544892333042179975631922610", 10)
	uY, _ := big.NewInt(0).SetString("27576123183859453704384360727380224739834659634660871190236925621255961659778", 10)
	U := curveField.MakeElement(uX, uY) // TODO: I cheat here and just construct U directly with values cribbed from pyUmbral
	return &Context{curveField, targetField, getMinValSha512(curveField), U, SECRET_BOX_KEY_SIZE}
}

func MakeSM2Context() *Context {
	curveField := MakeSM2()
	targetField := MakeZField(curveField.FieldOrder)
	uX, _ := big.NewInt(0).SetString("21988818426344374455592705741884136131748000592119723665854584937774395241572", 10)
	uY, _ := big.NewInt(0).SetString("31511737151452315797500599923801180801680237233295617230109528048179423546071", 10)
	U := curveField.MakeElement(uX, uY)
	return &Context{curveField, targetField, getMinValSha512(curveField), U, SECRET_SM4_KEY_SIZE}
}

func Aaaa() {

	cxt := MakeDefaultContext()
	//Alice
	privKeyAlice := GenPrivateKey(cxt)
	pubKeyAlice := privKeyAlice.GetPublicKey(cxt)
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
		//t.Errorf("Direct decryption failed")
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
		//t.Errorf("Re-encapsulated fragment decryption failed")
	}
	fmt.Println("true")
}
