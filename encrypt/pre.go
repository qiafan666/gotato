package CSCrypto

import (
	"crypto/sha512"
	"fmt"
	"github.com/tjfoc/gmsm/sm4"
	"golang.org/x/crypto/hkdf"
	"hash"
	"log"
	"math/big"
)

type Capsule struct {
	E *UmbralCurveElement
	V *UmbralCurveElement
	S *ZElement
}

// TODO: not complete implementation of Capsule or serialization
func (c *Capsule) toBytes() []byte {
	return append(
		append(c.E.toBytes(true), c.V.toBytes(true)...),
		BytesPadBigEndian(c.S.GetValue(), c.S.ElemField.LengthInBytes)...)
}

func (c *Capsule) verify(cxt *Context) bool {
	items := []([]byte){c.E.toBytes(true), c.V.toBytes(true)}
	h := hashToModInt(cxt, items)

	l := cxt.curveField.GetGen().MulScalar(c.S.GetValue())
	r := c.E.MulScalar(h.GetValue()).Add(&c.V.CurveElement)
	return l.IsValEqual(&r.PointLike)
}

func Encrypt(cxt *Context, pubKey *UmbralCurveElement, plainText []byte) ([]byte, *Capsule) {

	key, capsule := encapsulate(cxt, pubKey)

	capsuleBytes := capsule.toBytes()

	dem := MakeDEM(key)
	cypher := dem.encrypt(plainText, capsuleBytes)

	return cypher, capsule
}

func EncryptSM4(cxt *Context, pubKey *UmbralCurveElement, plainText []byte) ([]byte, *Capsule) {
	key, capsule := encapsulate(cxt, pubKey)
	dst, err := sm4.Sm4Ecb(key, plainText, true)
	if err != nil {
		return nil, nil
	}
	return dst, capsule
}

func DecryptDirect(cxt *Context, capsule *Capsule, privKey *UmbralFieldElement, cipherText []byte) []byte {

	key := decapDirect(cxt, privKey, capsule)
	dem := MakeDEM(key)

	capsuleBytes := capsule.toBytes()

	return dem.decrypt(cipherText, capsuleBytes)
}

func DecryptDirectSM4(cxt *Context, capsule *Capsule, privKey *UmbralFieldElement, cipherText []byte) []byte {
	key := decapDirect(cxt, privKey, capsule)
	dst, err := sm4.Sm4Ecb(key, cipherText, false)
	if err != nil {
		return nil
	}
	return dst
}

func DecryptFragments(cxt *Context, capsule *Capsule, reKeyFrags []*CFrag, privKey *UmbralFieldElement, origPubKey *UmbralCurveElement, cipherText []byte) []byte {
	key := openCapsule(cxt, privKey, origPubKey, capsule, reKeyFrags)
	dem := MakeDEM(key)

	capsuleBytes := capsule.toBytes()

	return dem.decrypt(cipherText, capsuleBytes)
}

func DecryptFragmentsSM4(cxt *Context, capsule *Capsule, reKeyFrags []*CFrag, privKey *UmbralFieldElement, origPubKey *UmbralCurveElement, cipherText []byte) []byte {
	key := openCapsule(cxt, privKey, origPubKey, capsule, reKeyFrags)
	dst, err := sm4.Sm4Ecb(key, cipherText, false)
	if err != nil {
		return nil
	}
	return dst
}

func hornerPolyEval(poly []*ModInt, x *ModInt) *ModInt {
	result := poly[0]
	for i := 1; i < len(poly); i++ {
		result = result.Mul(x).Add(poly[i])
	}
	return result
}

type KFrag struct {
	Id    *ModInt
	Rk    *ModInt
	XComp *UmbralCurveElement
	U1    *UmbralCurveElement
	Z1    *ModInt
	Z2    *ModInt
}
type CKFrag struct {
	ID     *ModInt
	RK     *ModInt
	XComp  *UmbralCurveElement
	XCompp *UmbralCurveElement
	U1     *UmbralCurveElement
	Z1     *ModInt
	Z2     *ModInt
}

func makeShamirPolyCoeffs(cxt *Context, coeff0 *ModInt, threshold int) []*ModInt {
	coeffs := make([]*ModInt, threshold-1)
	for i := range coeffs {
		coeffs[i] = MakeModIntRandom(cxt.GetOrder())
	}
	return append(coeffs, coeff0)
}

func GenerateCKFrags(cxt *Context, privA *UmbralFieldElement, pubB *UmbralCurveElement, pubProxy []*UmbralCurveElement, threshold int, numSplits int) []*CKFrag {
	kFrags := ReKeyGen(cxt, privA, pubB, threshold, numSplits)
	ckFrags := make([]*CKFrag, numSplits)
	for i := range kFrags {
		kFrag := kFrags[i]
		tmpPriv := GenPrivateKey(cxt)
		//tmpPub:=tmpPriv.GetPublicKey(cxt)
		calcPublicKey := cxt.curveField.GetGen().MulScalar(tmpPriv.GetValue()).Invert()
		tmpPub := &UmbralCurveElement{*calcPublicKey}
		ckFrags[i] = &CKFrag{kFrag.Id,
			kFrag.Rk,
			kFrag.XComp.Add(pubProxy[i].Mul(tmpPriv)),
			tmpPub,
			kFrag.U1,
			kFrag.Z1,
			kFrag.Z2}
	}
	return ckFrags
}
func GetKFrags(cxt *Context, privA []*UmbralFieldElement, ckFrags []*CKFrag) []*KFrag {
	kFrags := make([]*KFrag, len(ckFrags))
	for i := range kFrags {
		ckFrag := ckFrags[i]
		kFrags[i] = &KFrag{ckFrag.ID,
			ckFrag.RK,
			ckFrag.XComp.Add(ckFrag.XCompp.Mul(privA[i])),
			ckFrag.U1,
			ckFrag.Z1,
			ckFrag.Z2}
	}
	return kFrags
}

func ReKeyGen(cxt *Context, privA *UmbralFieldElement, pubB *UmbralCurveElement, threshold int, numSplits int) []*KFrag {

	pubA := privA.GetPublicKey(cxt)

	x := GenPrivateKey(cxt)
	xComp := x.GetPublicKey(cxt) // gen^x

	dhB := pubB.Mul(x) // pk_b^x

	// hash of:
	// . gen^x - where x is ephemeral
	// . gen^B - public key of B
	// . pk_b^x - so gen^(bx)
	d := hashToModInt(cxt, [][]byte{
		xComp.toBytes(true),
		pubB.toBytes(true),
		dhB.toBytes(true)})

	coeff0 := privA.Mul(d.Invert())

	coeffs := makeShamirPolyCoeffs(cxt, coeff0, threshold)

	kFrags := make([]*KFrag, numSplits)
	for i := range kFrags {
		id := MakeModIntRandom(cxt.GetOrder())
		rk := hornerPolyEval(coeffs, id)

		u1 := cxt.MulU(rk) // U^RK

		y := MakeModIntRandom(cxt.GetOrder())

		z1 := hashToModInt(cxt, [][]byte{
			cxt.MulGen(y).toBytes(true),                                    // gen^y
			BytesPadBigEndian(id.GetValue(), cxt.curveField.LengthInBytes), // TODO: ugly :/
			pubA.toBytes(true),
			pubB.toBytes(true),
			u1.toBytes(true),
			xComp.toBytes(true),
		})
		z2 := y.Sub(privA.Mul(z1))

		// kfrag is:
		// . ID - random element of Zq - input to shamir poly
		// . RK - result of shamir poly eval
		// . gen^x - used as input to d
		// . U^RK ??? why? U is A parameter and RK is known
		// . hash of (gen^y, ID, gen^A, gen^B, U^RK, gen^x)
		// . y (random Z element) - (privA * hash)
		kFrags[i] = &KFrag{id, rk, xComp, u1, z1, z2}
	}

	return kFrags
}

type CFrag struct {
	E1 *UmbralCurveElement
	V1 *UmbralCurveElement
	Id *ModInt
	X  *UmbralCurveElement
}

func ReEncapsulate(frag *KFrag, cap *Capsule) *CFrag {

	e1 := cap.E.MulInt(frag.Rk)
	v1 := cap.V.MulInt(frag.Rk)

	return &CFrag{e1, v1, frag.Id, frag.XComp}
}

func calcLPart(inId *ModInt, calcIdOrd int, ids []*ModInt) *ModInt {
	var result *ModInt
	if ids[calcIdOrd].IsValEqual(inId) {
		result = MI_ONE
	} else {
		div := ids[calcIdOrd].Sub(inId).Invert()
		result = ids[calcIdOrd].Mul(div)
	}
	if calcIdOrd == len(ids)-1 {
		return result
	} else {
		return calcLPart(inId, calcIdOrd+1, ids).Mul(result)
	}
}

func calcLambdaCoeff(inId *ModInt, selectedIds []*ModInt) *ModInt {
	switch len(selectedIds) {
	case 0:
		return nil
	case 1:
		if selectedIds[0].IsValEqual(inId) {
			return nil
		}
	}
	return calcLPart(inId, 0, selectedIds)
}

func reconstructSecret(inFrags []*CFrag) (*UmbralCurveElement, *UmbralCurveElement, *UmbralCurveElement) {

	if len(inFrags) == 0 {
		log.Panicf("Shouldn't call this with no fragments")
	}

	ids := make([]*ModInt, len(inFrags))
	for x, cf := range inFrags {
		ids[x] = cf.Id
	}

	var eFinal *UmbralCurveElement = nil
	var vFinal *UmbralCurveElement = nil

	for _, cf := range inFrags {
		lambda := calcLambdaCoeff(cf.Id, ids)
		e := cf.E1.MulInt(lambda)
		v := cf.V1.MulInt(lambda)
		eFinal = e.Add(eFinal)
		vFinal = v.Add(vFinal)
	}

	return eFinal, vFinal, inFrags[0].X
}

func kdf(keyElem *UmbralCurveElement, keySize int) []byte {

	keyMaster := hkdf.New(sha512.New, keyElem.toBytes(true), nil, nil)

	derivedKey := make([]byte, keySize)
	keyMaster.Read(derivedKey)

	return derivedKey
}

func encapsulate(cxt *Context, pubKey *UmbralCurveElement) ([]byte, *Capsule) {

	skR := GenPrivateKey(cxt)
	pkR := skR.GetPublicKey(cxt)

	skU := GenPrivateKey(cxt)
	pkU := skU.GetPublicKey(cxt)

	items := []([]byte){pkR.toBytes(true), pkU.toBytes(true)}
	h := hashToModInt(cxt, items)

	s := skU.Add(skR.Mul(h))
	sElem := cxt.targetField.NewElement(s.GetValue())

	sharedKey := pubKey.MulInt(skR.Add(skU.ModInt))

	symmetricKey := kdf(sharedKey, cxt.symKeySize)

	return symmetricKey, &Capsule{pkR, pkU, sElem}
}

func decapDirect(cxt *Context, privKey *UmbralFieldElement, capsule *Capsule) []byte {

	sharedKey := capsule.E.Add(capsule.V).MulInt(privKey.ModInt)
	key := kdf(sharedKey, cxt.symKeySize)

	if !capsule.verify(cxt) {
		log.Panicf("Capsule validation failed.") // TODO: not sure this should be A panic
	}

	return key
}

func decapReEncrypted(cxt *Context, targetPrivKey *UmbralFieldElement, origPublicKey *UmbralCurveElement, rec *ReEncCapsule) []byte {

	targetPubKey := targetPrivKey.GetPublicKey(cxt)

	// same computation as in SplitReKey except that the target private key is now known
	d := hashToModInt(cxt, [][]byte{
		rec.PointNI.toBytes(true),
		targetPubKey.toBytes(true),
		rec.PointNI.MulInt(targetPrivKey.ModInt).toBytes(true)})

	// from encapsulate: sharedKey := pubKey.MulScalar(skR.Add(skU.ModInt).GetValue())
	// shared_key = d * (e_prime + v_prime)
	sharedKey := rec.EPrime.Add(rec.VPrime).MulInt(d)

	symmetricKey := kdf(sharedKey, cxt.symKeySize)

	// checking...
	e := rec.OrigCap.E
	v := rec.OrigCap.V
	s := rec.OrigCap.S
	h := hashToModInt(cxt, [][]byte{
		e.toBytes(true),
		v.toBytes(true),
	})
	invD := d.Invert()

	// TODO: similar (?) to logic in verify(...)
	l := origPublicKey.MulInt(s.Mul(invD))
	r := rec.EPrime.MulInt(h).Add(rec.VPrime)
	if !l.IsValEqual(&r.PointLike) {
		log.Panicf("Failed decapulation check")
	}

	return symmetricKey
}

type ReEncCapsule struct {
	OrigCap *Capsule
	EPrime  *UmbralCurveElement
	VPrime  *UmbralCurveElement
	PointNI *UmbralCurveElement
}

func openCapsule(cxt *Context, targetPrivKey *UmbralFieldElement, origPublicKey *UmbralCurveElement, capsule *Capsule, reKeyFrags []*CFrag) []byte {

	ePrime, vPrime, pointNI := reconstructSecret(reKeyFrags)
	rec := &ReEncCapsule{capsule, ePrime, vPrime, pointNI}

	return decapReEncrypted(cxt, targetPrivKey, origPublicKey, rec)
}

func traceHash(h hash.Hash) {
	testDigest := h.Sum(nil)
	println(fmt.Sprintf("B'%x'", testDigest))
}

func hashToModInt(cxt *Context, items []([]byte)) *ModInt {

	createAndInitHash := func() hash.Hash {
		hasher := sha512.New()
		for _, item := range items {
			hasher.Write(item)
		}
		return hasher
	}

	i := int64(0)
	h := big.NewInt(0)
	for h.Cmp(cxt.minValSha512) < 0 {
		hasher := createAndInitHash()

		iBigEndianPadded := BytesPadBigEndian(big.NewInt(i), cxt.targetField.LengthInBytes)

		hasher.Write(iBigEndianPadded)
		hashDigest := hasher.Sum(nil)
		h = big.NewInt(0).SetBytes(hashDigest) // SetBytes assumes big-endian
		i += 1
	}

	res := CopyFrom(h.Mod(h, cxt.targetField.FieldOrder), true, cxt.targetField.FieldOrder)
	return res
}
