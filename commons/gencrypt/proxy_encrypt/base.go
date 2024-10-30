package proxy_encrypt

import (
	"fmt"
	"log"
	"math/big"
)

var ZERO = big.NewInt(0)
var ONE = big.NewInt(1)
var TWO = big.NewInt(2)
var THREE = big.NewInt(3)

// for validation purposes this special value is assumed to match any other modulus
var MOD_ANY *big.Int = nil

var MI_ZERO = MakeModInt(0, true, MOD_ANY)
var MI_ONE = MakeModInt(1, true, MOD_ANY)
var MI_TWO = MakeModInt(2, true, MOD_ANY)
var MI_THREE = MakeModInt(3, true, MOD_ANY)
var MI_FOUR = MakeModInt(4, true, MOD_ANY)
var MI_SEVEN = MakeModInt(7, true, MOD_ANY)
var MI_EIGHT = MakeModInt(8, true, MOD_ANY)

/*
ModInt is intended to represent the base level of integer modular math for field computations.
What may be A bit confusing (and I need to think about) is that I don't intend this to be A replacement for big.Int everywhere.
The full name here is more explicit: field.ModInt - that is, A large integer that is A component of A field, which implies/requires modular math.
*/
type ModInt struct {
	V      big.Int
	Frozen bool
	M      *big.Int
}

func MakeModInt(x int64, frozen bool, mod *big.Int) *ModInt {
	return &ModInt{*big.NewInt(x), frozen, mod}
}

func MakeModIntRandom(order *big.Int) *ModInt {
	return &ModInt{*GetRandomInt(order), true, order}
}

func MakeModIntWords(w []big.Word, frozen bool, mod *big.Int) *ModInt {
	bi := new(big.Int)
	bi.SetBits(w)
	return &ModInt{*bi, frozen, mod}
}

func MakeModIntStr(x string, base int, mod *big.Int) *ModInt {
	ret := big.Int{}
	ret.SetString(x, base)
	return &ModInt{ret, true, mod}
}

func (bi *ModInt) GetValue() *big.Int {
	return &bi.V
}

func (bi *ModInt) GetMod() *big.Int {
	return bi.M
}

func (bi *ModInt) Freeze() {
	bi.Frozen = true
	return
}

func (bi *ModInt) isZero() bool {
	return bi.V.Cmp(ZERO) == 0
}

func copyFromBytes(biBytes []byte, frozen bool, mod *big.Int) *ModInt {
	newModInt := new(ModInt)
	newModInt.V.SetBytes(biBytes)
	newModInt.Frozen = frozen
	newModInt.M = mod
	return newModInt
}

func CopyFrom(bi *big.Int, frozen bool, mod *big.Int) *ModInt {
	if bi == nil {
		return nil
	}
	newModInt := new(ModInt)
	newModInt.V.Set(bi)
	newModInt.Frozen = frozen
	newModInt.M = mod
	return newModInt
}

func (bi *ModInt) copyUnfrozen() *ModInt {
	if bi == nil {
		return nil
	}
	return CopyFrom(&bi.V, false, bi.M)
}

func (bi *ModInt) Copy() *ModInt {
	if bi == nil {
		return nil
	}
	return CopyFrom(&bi.V, bi.Frozen, bi.M)
}

func (bi *ModInt) setBytes(bytes []byte) {
	bi.V.SetBytes(bytes)
}

// TODO: how do we want these functions to behave WRT nil?
// also TODO: should we validate for modulus? right now no ...
func (bi *ModInt) IsValEqual(in *ModInt) bool {
	if bi == nil || in == nil {
		return false
	}
	return bi.V.Cmp(&in.V) == 0
}

func (bi *ModInt) modInternal(in *ModInt) *ModInt {
	if bi.Frozen {
		bi = bi.copyUnfrozen()
	}

	var mod = bi.M

	// common case should be that both ModInt's point to same big.Int modulus instance
	if bi.M != in.M {
		if in.M == MOD_ANY {
			mod = bi.M
		} else if bi.M == MOD_ANY {
			mod = in.M
		} else if bi.M.Cmp(in.M) == 0 {
			mod = bi.M
		}
	}

	if mod == MOD_ANY {
		log.Panicf("Cannot perform mod arithmetic with unspecified or inconsistent modulo")
	}

	bi.V.Mod(&bi.V, mod)
	return bi
}

func (bi *ModInt) Add(in *ModInt) *ModInt {
	if bi.Frozen {
		bi = bi.copyUnfrozen()
	}
	bi.V.Add(&bi.V, &in.V)
	return bi.modInternal(in)
}

func (bi *ModInt) Sub(in *ModInt) *ModInt {
	if bi.Frozen {
		bi = bi.copyUnfrozen()
	}
	bi.V.Sub(&bi.V, &in.V)
	return bi.modInternal(in)
}

func (bi *ModInt) Mul(in *ModInt) *ModInt {
	if bi.Frozen {
		bi = bi.copyUnfrozen()
	}
	if in.IsValEqual(MI_ONE) {
		return bi
	}
	bi.V.Mul(&bi.V, &in.V)
	return bi.modInternal(in)
}

func (bi *ModInt) Square() *ModInt {
	if bi.Frozen {
		bi = bi.copyUnfrozen()
	}
	bi.V.Mul(&bi.V, &bi.V)
	return bi.modInternal(bi)
}

func (bi *ModInt) isSquare() bool {
	if bi.IsValEqual(MI_ZERO) {
		return true
	}
	return big.Jacobi(&bi.V, bi.M) == 1
}

func (bi *ModInt) sqrt() *ModInt {
	// Int.ModSqrt implements  Tonelli-Shanks and also A more optimal version when modIn = 3 mod 4
	// UGH! need to work around this bug: https://github.com/golang/go/issues/22265
	// for now always Copy
	calc := bi.copyUnfrozen()
	// TODO validate mod ? ie non-nil?
	calc.V.ModSqrt(&bi.V, bi.M)
	return calc
}

func (bi *ModInt) Invert() *ModInt {
	if bi.Frozen {
		bi = bi.copyUnfrozen()
	}
	bi.V.ModInverse(&bi.V, bi.M)
	return bi
}

func (bi *ModInt) Negate() *ModInt {
	if bi.isZero() {
		return MI_ONE.copyUnfrozen()
	}
	return CopyFrom(bi.M, false, bi.M).Sub(bi)
}

func (bi *ModInt) Pow(in *ModInt) *ModInt {
	if bi.Frozen {
		bi = bi.copyUnfrozen()
	}
	bi.V.Exp(&bi.V, &in.V, bi.M)
	return bi
}

func (bi *ModInt) String() string {
	return bi.V.String()
	// return bi.V.Text(16)
}

type Field interface{}

/*
type PointField interface {
	MakeElement(x *ModInt, y *ModInt) PointElement
}
*/

type PowElement interface {
	String() string
	CopyPow() PowElement
	MakeOnePow() PowElement
	MulPow(PowElement) PowElement
}

/*
type PointElement interface {
	String() string
	X() *ModInt
	Y() *ModInt
	NegateY() PointElement
	Invert() PointElement
	MulPoint(PointElement) PointElement
	Add(PointElement) PointElement
	Sub(PointElement) PointElement
	Square() PointElement
	Pow(*ModInt) PointElement
	IsValEqual(PointElement) bool
}
*/

type BaseField struct {
	LengthInBytes int
	FieldOrder    *big.Int
}

type PointLike struct {
	DataX *ModInt
	DataY *ModInt
}

func (p *PointLike) IsInf() bool {
	return p.DataX == nil && p.DataY == nil
}

// TODO: get at base type name ?
func (p *PointLike) String() string {
	if p.IsInf() {
		return "[INF]"
	}
	return fmt.Sprintf("PointLike: [%s,\n%s]", p.DataX.String(), p.DataY.String())
}

func (p *PointLike) freeze() {
	p.DataX.Freeze()
	p.DataY.Freeze()
}

func (p *PointLike) frozen() bool {
	return p.DataX.Frozen && p.DataY.Frozen
}

func (p *PointLike) X() *ModInt {
	return p.DataX
}

func (p *PointLike) Y() *ModInt {
	return p.DataY
}

func (p *PointLike) IsValEqual(elemIn *PointLike) bool {
	return p.DataX.IsValEqual(elemIn.X()) && p.DataY.IsValEqual(elemIn.Y())
}

func MakePointFromBytes(pointBytes []byte, targetField *BaseField) *PointLike {

	if len(pointBytes) != targetField.LengthInBytes*2 {
		log.Panicf("Point byte data must have length of 2X target field: got %v, expect %v", len(pointBytes), targetField.LengthInBytes*2)
	}

	xBytes := pointBytes[:targetField.LengthInBytes]
	yBytes := pointBytes[targetField.LengthInBytes:]

	DataX := new(ModInt)
	DataX.setBytes(xBytes)
	DataX.M = targetField.FieldOrder

	DataY := new(ModInt)
	DataY.setBytes(yBytes)
	DataY.M = targetField.FieldOrder

	return &PointLike{DataX, DataY}
}

func powWindow(base PowElement, exp *big.Int) PowElement {

	// note: does not mutate base
	result := base.MakeOnePow()

	if exp.Sign() == 0 {
		return result
	}

	k := optimalPowWindowSize(exp)
	lookups := buildPowWindow(k, base)

	word := uint(0)
	wordBits := uint(0)

	inWord := false
	for s := exp.BitLen() - 1; s >= 0; s-- {
		result = result.MulPow(result)

		bit := exp.Bit(s)

		if !inWord && bit == 0 {
			continue
		}

		if !inWord {
			inWord = true
			word = 1
			wordBits = 1
		} else {
			word = (word << 1) + bit
			wordBits++
		}

		if wordBits == k || s == 0 {
			result = result.MulPow((*lookups)[word])
			inWord = false
		}
	}

	return result
}

func optimalPowWindowSize(exp *big.Int) uint {

	expBits := exp.BitLen()

	switch {
	case expBits > 9065:
		return 8
	case expBits > 3529:
		return 7
	case expBits > 1324:
		return 6
	case expBits > 474:
		return 5
	case expBits > 157:
		return 4
	case expBits > 47:
		return 3
	default:
		return 2
	}
}

func buildPowWindow(k uint, base PowElement) *[]PowElement {

	if k < 1 {
		return nil
	}

	lookupSize := 1 << k
	lookups := make([]PowElement, lookupSize)

	// MakeOnePow copies ...
	lookups[0] = base.MakeOnePow()
	for x := 1; x < lookupSize; x++ {
		newLookup := lookups[x-1].CopyPow()
		lookups[x] = newLookup.MulPow(base)
	}

	return &lookups
}
