package CSCrypto

import (
	"fmt"
	"github.com/tjfoc/gmsm/sm4"
	"os"
)

const (
	Encrypt256k1 = "256k1"
	EncryptSm2   = "sm2"
)

type ProxyEncrypt struct {
	EncryptMethod    string //加密方式 256k1或sm2
	PrivateKey       *UmbralFieldElement
	PublicKey        *UmbralCurveElement
	PrivateKeyString string
	PublicKeyString  string
}

// NewProxyEncrypt 不传默认sm2
func NewProxyEncrypt(encryptMethod ...string) *ProxyEncrypt {
	var privateKeyString, publicKeyString, method string
	var privateKey *UmbralFieldElement
	var publicKey *UmbralCurveElement

	if len(encryptMethod) > 1 {
		panic("encryptMethod error")
	}
	if len(encryptMethod) > 0 {
		method = encryptMethod[0]
		if encryptMethod[0] == Encrypt256k1 {
			privateKey = GenPrivateKey(MakeDefaultContext())
			publicKey = privateKey.GetPublicKey(MakeDefaultContext())
			str := StructToString(publicKey)
			fmt.Fprintf(os.Stdout, str)
			privateKeyString = StructToString(privateKey)
			publicKeyString = StructToString(publicKey)
		} else if encryptMethod[0] == EncryptSm2 {
			privateKey = GenPrivateKey(MakeSM2Context())
			publicKey = privateKey.GetPublicKey(MakeSM2Context())
			privateKeyString = SM2PrivateToString(privateKey)
			publicKeyString = SM2PublicToString(publicKey)
		} else {
			privateKey = GenPrivateKey(MakeSM2Context())
			publicKey = privateKey.GetPublicKey(MakeSM2Context())
			privateKeyString = SM2PrivateToString(privateKey)
			publicKeyString = SM2PublicToString(publicKey)
		}
	} else {
		method = EncryptSm2
		privateKey = GenPrivateKey(MakeSM2Context())
		publicKey = privateKey.GetPublicKey(MakeSM2Context())
		privateKeyString = SM2PrivateToString(privateKey)
		publicKeyString = SM2PublicToString(publicKey)
	}

	return &ProxyEncrypt{
		EncryptMethod:    method,
		PrivateKey:       privateKey,
		PublicKey:        publicKey,
		PrivateKeyString: privateKeyString,
		PublicKeyString:  publicKeyString,
	}
}

func (proxyEncrypt *ProxyEncrypt) Sign(msg []byte) (*UmbralFieldElement, *UmbralFieldElement) {

	var cxt *Context
	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		cxt = MakeDefaultContext()

	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		cxt = MakeSM2Context()
	} else {
		return nil, nil
	}

	h := GenPrivateKeyFromMsg(cxt, msg)
	k := &UmbralFieldElement{*cxt.targetField.NewElement(ONE)} //GenPrivateKey(cxt)
	kG := k.GetPublicKey(cxt)
	_r := kG.DataX
	e := cxt.targetField.NewElement(_r.GetValue())
	r := &UmbralFieldElement{*e}
	_s := proxyEncrypt.PrivateKey.Mul(r.ModInt).Add(h.ModInt).Mul(k.Invert())
	e = cxt.targetField.NewElement(_s.GetValue())
	s := &UmbralFieldElement{*e}
	return r, s
}

func (proxyEncrypt *ProxyEncrypt) Verify(r *UmbralFieldElement, s *UmbralFieldElement, msg []byte) bool {

	var cxt *Context
	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		cxt = MakeDefaultContext()

	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		cxt = MakeSM2Context()
	} else {
		return false
	}

	h := GenPrivateKeyFromMsg(cxt, msg)
	s_1 := s.Invert()
	h_s1 := h.Mul(s_1).GetValue()
	r_s1 := r.Mul(s_1).GetValue()
	P_1 := cxt.curveField.GetGen().MulScalar(h_s1)
	P_2 := proxyEncrypt.PublicKey.MulScalar(r_s1)
	R2 := P_1.Add(P_2)
	//fmt.Println(r.GetValue(), R2.DataX.GetValue())
	return r.GetValue().Cmp(R2.DataX.GetValue()) == 0
}

func (proxyEncrypt *ProxyEncrypt) Encrypt(plainText []byte) ([]byte, *Capsule) {

	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		key, capsule := encapsulate(MakeDefaultContext(), proxyEncrypt.PublicKey)
		capsuleBytes := capsule.toBytes()
		dem := MakeDEM(key)
		cypher := dem.encrypt(plainText, capsuleBytes)
		return cypher, capsule
	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		key, capsule := encapsulate(MakeSM2Context(), proxyEncrypt.PublicKey)
		dst, err := sm4.Sm4Ecb(key, plainText, true)
		if err != nil {
			return nil, nil
		}
		return dst, capsule
	} else {
		return []byte{}, nil
	}
}

func (proxyEncrypt *ProxyEncrypt) Decrypt(capsule *Capsule, cipherText []byte) []byte {

	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		key := decapDirect(MakeDefaultContext(), proxyEncrypt.PrivateKey, capsule)
		dem := MakeDEM(key)
		capsuleBytes := capsule.toBytes()
		return dem.decrypt(cipherText, capsuleBytes)
	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		key := decapDirect(MakeSM2Context(), proxyEncrypt.PrivateKey, capsule)
		dst, err := sm4.Sm4Ecb(key, cipherText, false)
		if err != nil {
			return nil
		}
		return dst
	} else {
		return []byte{}
	}
}

func (proxyEncrypt *ProxyEncrypt) DecryptFragments(capsule *Capsule, reKeyFrags []*CFrag, origPubKey *UmbralCurveElement, cipherText []byte) []byte {

	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		key := openCapsule(MakeDefaultContext(), proxyEncrypt.PrivateKey, origPubKey, capsule, reKeyFrags)
		dem := MakeDEM(key)
		capsuleBytes := capsule.toBytes()
		return dem.decrypt(cipherText, capsuleBytes)
	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		key := openCapsule(MakeSM2Context(), proxyEncrypt.PrivateKey, origPubKey, capsule, reKeyFrags)
		dst, err := sm4.Sm4Ecb(key, cipherText, false)
		if err != nil {
			return nil
		}
		return dst
	} else {
		return []byte{}
	}
}
func (proxyEncrypt *ProxyEncrypt) Pri2String() string {

	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		return StructToString(proxyEncrypt.PrivateKey)
	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		return SM2PrivateToString(proxyEncrypt.PrivateKey)
	} else {
		return ""
	}
}

func (proxyEncrypt *ProxyEncrypt) Pub2String() string {

	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		return StructToString(proxyEncrypt.PublicKey)
	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		return SM2PublicToString(proxyEncrypt.PublicKey)
	} else {
		return ""
	}
}

func (proxyEncrypt *ProxyEncrypt) String2Pri(pri string) *UmbralFieldElement {

	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		var obj = new(UmbralFieldElement)
		StringToStruct(pri, obj)
		return obj
	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		return SM2StringToPrivate(pri)
	} else {
		return nil
	}
}

func (proxyEncrypt *ProxyEncrypt) String2Pub(pub string) *UmbralCurveElement {

	if proxyEncrypt.EncryptMethod == Encrypt256k1 {
		var obj = new(UmbralCurveElement)
		StringToStruct(pub, obj)
		return obj
	} else if proxyEncrypt.EncryptMethod == EncryptSm2 {
		return SM2StringToPublic(pub)
	} else {
		return nil
	}
}

func EncryptByMethod(encryptMethod string, pubKey *UmbralCurveElement, plainText []byte) ([]byte, *Capsule) {

	if encryptMethod == Encrypt256k1 {
		key, capsule := encapsulate(MakeDefaultContext(), pubKey)
		capsuleBytes := capsule.toBytes()
		dem := MakeDEM(key)
		cypher := dem.encrypt(plainText, capsuleBytes)
		return cypher, capsule
	} else if encryptMethod == EncryptSm2 {
		key, capsule := encapsulate(MakeSM2Context(), pubKey)
		dst, err := sm4.Sm4Ecb(key, plainText, true)
		if err != nil {
			return nil, nil
		}
		return dst, capsule
	} else {
		return []byte{}, nil
	}
}

func DecryptByMethod(encryptMethod string, capsule *Capsule, privKey *UmbralFieldElement, cipherText []byte) []byte {

	if encryptMethod == Encrypt256k1 {
		key := decapDirect(MakeDefaultContext(), privKey, capsule)
		dem := MakeDEM(key)
		capsuleBytes := capsule.toBytes()
		return dem.decrypt(cipherText, capsuleBytes)
	} else if encryptMethod == EncryptSm2 {
		key := decapDirect(MakeSM2Context(), privKey, capsule)
		dst, err := sm4.Sm4Ecb(key, cipherText, false)
		if err != nil {
			return nil
		}
		return dst
	} else {
		return []byte{}
	}
}

func DecryptFragmentsByMethod(encryptMethod string, capsule *Capsule, reKeyFrags []*CFrag, privKey *UmbralFieldElement, origPubKey *UmbralCurveElement, cipherText []byte) []byte {

	if encryptMethod == Encrypt256k1 {
		key := openCapsule(MakeDefaultContext(), privKey, origPubKey, capsule, reKeyFrags)
		dem := MakeDEM(key)
		capsuleBytes := capsule.toBytes()
		return dem.decrypt(cipherText, capsuleBytes)
	} else if encryptMethod == EncryptSm2 {
		key := openCapsule(MakeSM2Context(), privKey, origPubKey, capsule, reKeyFrags)
		dst, err := sm4.Sm4Ecb(key, cipherText, false)
		if err != nil {
			return nil
		}
		return dst
	} else {
		return []byte{}
	}
}
