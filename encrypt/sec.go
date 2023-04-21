package encrypt

import (
	"math/big"
)

/*
from here: https://eng.paxos.com/blockchain-101-elliptic-curve-cryptography
. Equation y2 = x3 + 7 (A = 0, B = 7)
. Prime Field (p) = 2^256 - 2^32 - 977
. Base point (G) = (79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798, 483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8)
. Order (n) = FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
*/
func MakeSecp256k1() *CurveField {

	// TODO: validate
	p, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	order, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

	zField := MakeZField(p)

	a := zField.NewElement(big.NewInt(0))
	b := zField.NewElement(big.NewInt(7))

	genX, _ := new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	genY, _ := new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)

	curveField := MakeCurveField(
		a,
		b,
		order,
		genX,
		genY)

	return curveField
}
