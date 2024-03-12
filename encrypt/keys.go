package CSCrypto

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
)

type UmbralFieldElement struct {
	ZElement
}

type UmbralCurveElement struct {
	CurveElement
}

// TODO: hmmm... need to think about this ...
func toUmbralBytes(elem *CurveElement, compressed bool, keySize int) []byte {
	if compressed {
		yBit := big.NewInt(0).And(elem.Y().GetValue(), ONE)
		yBit = yBit.Add(yBit, TWO)
		return append(yBit.Bytes(), BytesPadBigEndian(elem.X().GetValue(), keySize)...)
	} else {
		data := make([]byte, 1)
		data[0] = 0x04
		data = append(data, BytesPadBigEndian(elem.X().GetValue(), keySize)...)
		data = append(data, BytesPadBigEndian(elem.Y().GetValue(), keySize)...)
		return data
	}
}

// TODO: currently implementing this on UmbralCurveElement - as opposed to CurveElement - because pyUmbral has A very specific way of serializing
func (key *UmbralCurveElement) toBytes(compressed bool) []byte {
	return toUmbralBytes(&key.CurveElement, true, key.ElemParams.GetTargetField().LengthInBytes)
}

func (key *UmbralCurveElement) MulInt(mi *ModInt) *UmbralCurveElement {
	if mi == nil {
		return key
	}
	return &UmbralCurveElement{*key.MulScalar(mi.GetValue())}
}

func (key *UmbralCurveElement) Add(in *UmbralCurveElement) *UmbralCurveElement {
	if in == nil {
		return key
	}
	return &UmbralCurveElement{*key.CurveElement.Add(&in.CurveElement)}
}

// UmbralFieldElement
func GenPrivateKeyFromMsg(cxt *Context, msg []byte) *UmbralFieldElement {
	z := new(big.Int)
	z.SetBytes(msg)
	e := cxt.targetField.NewElement(z)
	return &UmbralFieldElement{*e}
}

func GenPrivateKey(cxt *Context) *UmbralFieldElement {
	randoKey := GetRandomInt(cxt.targetField.FieldOrder)
	e := cxt.targetField.NewElement(randoKey)
	return &UmbralFieldElement{*e}
}

func MakePrivateKey(cxt *Context, mi *ModInt) *UmbralFieldElement {
	e := cxt.targetField.NewElement(mi.GetValue())
	return &UmbralFieldElement{*e}
}

func (key *UmbralFieldElement) GetPublicKey(cxt *Context) *UmbralCurveElement {
	calcPublicKey := cxt.curveField.GetGen().MulScalar(key.GetValue())
	return &UmbralCurveElement{*calcPublicKey}
}

//
//// UmbralCurveElement
//
//func (pk *UmbralCurveElement) Mul(sk *UmbralFieldElement) *UmbralCurveElement {
//	return &UmbralCurveElement{*pk.MulScalar(sk.GetValue())}
//}
//
////公私钥、capsule， ckfrags
//func Struct2Bytes(c interface{}) []byte{
//	//testType(c)
//	var b bytes.Buffer
//	encoder := gob.NewEncoder(&b)
//	err := encoder.Encode(c)
//	if err != nil {
//		fmt.Println(err.Error())
//		return nil
//	}
//	return b.Bytes()
//}
//
//func Bytes2Struct(data []byte, obj interface{}){
//	decoder := gob.NewDecoder(bytes.NewReader(data))
//	err := decoder.Decode(obj)
//	if err != nil {
//		fmt.Println(err.Error())
//		obj=nil
//	}
//}
//
////公私钥、capsule， ckfrags
//func StructToString(o interface{}) string {
//	b:=Struct2Bytes(o)
//	return hex.EncodeToString(b)
//}
//func StringToStruct(s string, obj interface{}) {
//	b,_:= hex.DecodeString(s)
//	Bytes2Struct(b,obj)
//}

// UmbralCurveElement

func (pk *UmbralCurveElement) Mul(sk *UmbralFieldElement) *UmbralCurveElement {
	return &UmbralCurveElement{*pk.MulScalar(sk.GetValue())}
}

var priPrefix = "2eff8103010112556d6272616c4669656c64456c656d656e7401ff8200010101085a456c656d656e7401ff8400000031ff83030101085a456c656d656e7401ff840001020109456c656d4669656c6401ff860001064d6f64496e7401ff8c00000033ff85030101065a4669656c6401ff860001020109426173654669656c6401ff8800010a54776f496e766572736501ff8c00000039ff8703010109426173654669656c6401ff88000102010d4c656e677468496e4279746573010400010a4669656c644f7264657201ff8a0000000aff89050102ff8e0000002dff8b030101064d6f64496e7401ff8c00010301015601ff8a00010646726f7a656e01020001014d01ff8a000000ffc2ff820101010140012102fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd036414100010121027fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a10101012102fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141000001012102"
var priPostfix = "0101012102fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141000000"
var pubPreFix = "32ff8103010112556d6272616c4375727665456c656d656e7401ff82000101010c4375727665456c656d656e7401ff8400000039ff830301010c4375727665456c656d656e7401ff84000102010a456c656d506172616d7301ff86000109506f696e744c696b6501ff9200000036ff850301010b4375727665506172616d7301ff860001030109426173654669656c6401ff880001014101ff8c0001014201ff8c00000039ff8703010109426173654669656c6401ff88000102010d4c656e677468496e4279746573010400010a4669656c644f7264657201ff8a0000000aff89050102ff9400000031ff8b030101085a456c656d656e7401ff8c0001020109456c656d4669656c6401ff8e0001064d6f64496e7401ff9000000033ff8d030101065a4669656c6401ff8e0001020109426173654669656c6401ff8800010a54776f496e766572736501ff900000002dff8f030101064d6f64496e7401ff9000010301015601ff8a00010646726f7a656e01020001014d01ff8a0000002dff9103010109506f696e744c696b6501ff920001020105446174615801ff90000105446174615901ff90000000fe0204ff8201010101ff80012102fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141000101010140012102fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f00010121027fffffffffffffffffffffffffffffffffffffffffffffffffffffff7ffffe180101012102fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f0000010101020101012102fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f00000101010140012102fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f00010121027fffffffffffffffffffffffffffffffffffffffffffffffffffffff7ffffe180101012102fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f000001010202070101012102fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f0000000101012102"

// var pubPostFix = "0101012102fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f00000000"
var pubPostFix = "0101012102fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

// 公私钥、capsule， ckfrags
func Struct2Bytes(c interface{}) []byte {
	//testType(c)
	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	err := encoder.Encode(c)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	_, ok := c.(*UmbralFieldElement)
	_, ok2 := c.(*UmbralCurveElement)
	if ok {
		preb, _ := hex.DecodeString(priPrefix)
		postb, _ := hex.DecodeString(priPostfix)
		ab := b.Bytes()
		len1 := len([]byte(preb))
		//fmt.Println(len(ab),len1, len(postb))
		return ab[len1 : len(ab)-len(postb)]
	} else if ok2 {
		preb, _ := hex.DecodeString(pubPreFix)
		postb, _ := hex.DecodeString(pubPostFix)
		ab := b.Bytes()
		len1 := len([]byte(preb))
		return ab[len1 : len(ab)-len(postb)]
	}
	return b.Bytes()
}

func Bytes2Struct(data []byte, obj interface{}) {
	_, ok := obj.(*UmbralFieldElement)
	_, ok2 := obj.(*UmbralCurveElement)
	b1 := new(bytes.Buffer)
	if ok {
		preb, _ := hex.DecodeString(priPrefix)
		postb, _ := hex.DecodeString(priPostfix)
		b1.Write(preb)
		b1.Write(data)
		b1.Write(postb)
	} else if ok2 {
		preb, _ := hex.DecodeString(pubPreFix)
		postb, _ := hex.DecodeString(pubPostFix)
		b1.Write(preb)
		b1.Write(data)
		b1.Write(postb)
	} else {
		b1.Write(data)
	}
	decoder := gob.NewDecoder(bytes.NewReader(b1.Bytes()))
	err := decoder.Decode(obj)
	if err != nil {
		fmt.Println(err.Error())
		obj = nil
	}
}

// 公私钥、capsule， ckfrags
func StructToString(o interface{}) string {
	b := Struct2Bytes(o)
	return hex.EncodeToString(b)
}
func StringToStruct(s string, obj interface{}) {
	b, err := hex.DecodeString(s)
	if err != nil {
		fmt.Println("字符转结构体解码出错" + err.Error())
	}
	Bytes2Struct(b, obj)
}

func SM2PrivateToString(p *UmbralFieldElement) string {
	return p.GetValue().Text(16)
}

func SM2StringToPrivate(s string) *UmbralFieldElement {
	v, ok := new(big.Int).SetString(s, 16)
	if !ok {
		return nil
	}
	e := MakeSM2Context().targetField.NewElement(v)
	return &UmbralFieldElement{*e}
}

func SM2PublicToString(p *UmbralCurveElement) string {
	var buf []byte
	yp := getLastBit(p.DataY.GetValue())
	buf = append(buf, p.DataX.GetValue().Bytes()...)
	if n := len(p.DataX.GetValue().Bytes()); n < 32 {
		buf = append(zeroByteSlice()[:(32-n)], buf...)
	}
	buf = append([]byte{byte(yp)}, buf...)
	return hex.EncodeToString(buf)
}

func SM2StringToPublic(s string) *UmbralCurveElement {
	byt, err := hex.DecodeString(s)
	if err != nil {
		return nil
	}
	pub := sm2.Decompress(byt)
	curElement := MakeSM2Context().curveField.MakeElement(pub.X, pub.Y)
	return &UmbralCurveElement{*curElement}
}

func getLastBit(a *big.Int) uint {
	return a.Bit(0)
}

// 32byte
func zeroByteSlice() []byte {
	return []byte{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	}
}
