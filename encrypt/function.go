package CSCrypto

import (
	"github.com/tjfoc/gmsm/sm4"
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
