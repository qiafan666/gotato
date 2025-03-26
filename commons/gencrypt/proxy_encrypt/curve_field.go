package proxy_encrypt

import (
	"log"
	"math/big"
)

type CurveField struct {
	CurveParams
	cofactor   *big.Int      // TODO: do we need this ...?
	gen        *CurveElement // TODO: not sure here...
	genNoCofac *CurveElement // TODO: don't need this ...?
}

type CurveParams struct {
	BaseField
	A *ZElement
	B *ZElement
}

type CurveElement struct {
	ElemParams *CurveParams
	PointLike
}

// CurveField

// TODO: JPBC (PBC?) handles case w/o bytes and cofactor
func (field *CurveField) initGenFromBytes(genNoCofacBytes []byte) {
	if genNoCofacBytes == nil {
		return
	}
	newGenNoCoFac := field.MakeElementFromBytes(genNoCofacBytes)
	field.genNoCofac = newGenNoCoFac
	field.gen = field.genNoCofac.MulScalar(field.cofactor)
	if !field.gen.isValid() {
		panic("Curve field generator needs to be valid")
	}
}

func (field *CurveField) GetGen() *CurveElement {
	return field.gen
}

func (curveParams *CurveParams) GetTargetField() *ZField {
	return curveParams.A.ElemField
}

func (field *CurveField) MakeElementFromBytes(elemBytes []byte) *CurveElement {

	pnt := MakePointFromBytes(elemBytes, &field.GetTargetField().BaseField)

	elem := &CurveElement{&field.CurveParams, *pnt}

	// needs to be Frozen before validation
	elem.freeze()
	if !elem.isValid() {
		elem.setInf()
	}
	return elem
}

// general curve is y^2 = x^3 + ax + B
func (params *CurveParams) calcYSquared(xIn *ModInt) *ModInt {
	if !xIn.Frozen {
		panic("xIn needs to be Frozen")
	}
	validateModulo(params.GetTargetField().FieldOrder, xIn.M)
	return xIn.Square().Add(params.A.ModInt).Mul(xIn).Add(params.B.ModInt)
}

// this function constructs A point on the curve from the input hash-derived bytes.
// since the input is assumed to be random when we use it as an initial X value it is not guaranteed to lie on the curve
// therefore - unlike MakeElementFromX - we iterate in A stable way to find A value that does satisfy the curve equation
// the size of the hash must be such that we can guarantee that its value as an integer is less than our target order
func (field *CurveField) MakeElementFromHash(h []byte) *CurveElement {
	maxSafeBytes := field.GetTargetField().LengthInBytes - 1
	if len(h) > maxSafeBytes {
		log.Panicf("Cannot construct point from hash when byte length exceeds field capacity: max bytes %v, got %v", maxSafeBytes, len(h))
	}
	calcX := copyFromBytes(h, true, field.GetTargetField().FieldOrder)

	calcY2 := MI_ONE
	gotIt := false
	for !gotIt {
		calcY2 = field.calcYSquared(calcX)
		if calcY2.isSquare() {
			gotIt = true
		} else {
			calcX = calcX.Square().Add(MI_ONE)
			calcX.Freeze()
		}
	}

	calcY := calcY2.sqrt()
	if calcY.V.Sign() < 0 {
		calcY = calcY.Negate()
	}

	elem := &CurveElement{&field.CurveParams, PointLike{calcX, calcY2.sqrt()}}
	elem.freeze()
	if field.cofactor != nil {
		elem = elem.MulScalar(field.cofactor)
	}

	return elem
}

func (field *CurveField) MakeElement(x *big.Int, y *big.Int) *CurveElement {
	copyX := CopyFrom(x, true, field.GetTargetField().FieldOrder)
	copyY := CopyFrom(y, true, field.GetTargetField().FieldOrder)
	elem := CurveElement{&field.CurveParams, PointLike{copyX, copyY}}
	elem.freeze()
	return &elem
}

// TODO: needs to account for sign
func (field *CurveField) MakeElementFromX(x *big.Int) *CurveElement {

	copyX := CopyFrom(x, true, field.GetTargetField().FieldOrder)
	calcY2 := field.calcYSquared(copyX)
	if !calcY2.isSquare() {
		log.Panicf("Expected to calculate square: value %s", calcY2.String())
	}
	DataY := calcY2.sqrt()

	elem := CurveElement{&field.CurveParams, PointLike{copyX, DataY}}
	elem.freeze()
	return &elem
}

func (field *CurveField) newElementFromStrings(xStr string, yStr string) *CurveElement {
	targetOrder := field.GetTargetField().FieldOrder
	return &CurveElement{&field.CurveParams,
		PointLike{MakeModIntStr(xStr, 10, targetOrder), MakeModIntStr(yStr, 10, targetOrder)}}
}

func getLengthInBytes(field *CurveField) int {
	return field.GetTargetField().LengthInBytes * 2
}

func MakeCurveField(
	a *ZElement,
	b *ZElement,
	order *big.Int,
	genX *big.Int,
	genY *big.Int) *CurveField {

	field := new(CurveField)
	field.A = a
	field.B = b
	field.FieldOrder = order
	field.LengthInBytes = getLengthInBytes(field)

	field.gen = field.MakeElement(genX, genY)

	if !field.gen.isValid() {
		panic("Curve field generator needs to be valid")
	}

	field.cofactor = nil // TODO: not sure if we need / want this...

	return field
}

// TODO: need to reconcile this and the other make function - not sure I both ...?
// make minimal field for testing purposes - TODO: might need A generator?
func makeTestCurveField(a *big.Int, b *big.Int, r *big.Int, q *big.Int) *CurveField {

	zfield := MakeZField(q)

	cfield := new(CurveField)
	cfield.A = zfield.NewElement(a)
	cfield.B = zfield.NewElement(b)
	cfield.FieldOrder = r
	cfield.LengthInBytes = getLengthInBytes(cfield)

	return cfield
}

// CurveElement

// TODO: Make function?

// var _ PointElement = (*CurveElement)(nil)
var _ IPowElement = (*CurveElement)(nil)

func (elem *CurveElement) getTargetOrder() *big.Int {
	return elem.ElemParams.GetTargetField().FieldOrder
}

func (elem *CurveElement) NegateY() *CurveElement {
	if elem.IsInf() {
		return &CurveElement{elem.ElemParams, PointLike{nil, nil}}
	}
	elem.PointLike.freeze() // make sure we're Frozen
	yNeg := elem.DataY.Negate()
	return &CurveElement{elem.ElemParams, PointLike{elem.DataX, yNeg}}
}

func (elem *CurveElement) Invert() *CurveElement {
	if elem.IsInf() {
		return elem
	}
	elem.DataY = elem.DataY.Negate()
	elem.DataY.Freeze()
	return elem
}

func (elem *CurveElement) Square() *CurveElement {
	// TODO !?
	return nil
}

func (elem *CurveElement) Add(elemIn *CurveElement) *CurveElement {
	return elem.MulPoint(elemIn)
}

func (elem *CurveElement) Sub(_ *CurveElement) *CurveElement {
	return nil // TODO!?
}

/*
func (elem *CurveElement) IsInf() bool {
	return elem.DataY == nil && elem.DataY == nil
}
*/

func (elem *CurveElement) setInf() {
	elem.DataX = nil
	elem.DataY = nil
}

// don't return elem to emphasize that call mutates elem
func (elem *CurveElement) freeze() {
	if elem.IsInf() {
		return // already Frozen by def
	}
	elem.PointLike.freeze()
	return
}

func (elem *CurveElement) frozen() bool {
	if elem.IsInf() {
		return true
	}
	return elem.PointLike.frozen()
}

func (elem *CurveElement) MulScalar(n *big.Int) *CurveElement {
	result := powWindow(elem, n).(*CurveElement)
	result.freeze()
	return result
}

func (elem *CurveElement) PowZn(in *big.Int) *CurveElement {
	result := powWindow(elem, in).(*CurveElement)
	result.freeze()
	return result
}

func (elem *CurveElement) Pow(in *ModInt) *CurveElement {
	return elem.PowZn(&in.V)
}

func validateModulo(mod1 *big.Int, mod2 *big.Int) {
	// TODO: this is intentionally pointer comparison because we expect the ModInt M's to point to the same object
	// need to think about this tho ...
	if mod1 == nil || mod1 != mod2 {
		log.Panicf("Field components must have valid and equal modulo")
	}
}

func (elem *CurveElement) isValid() bool {

	if elem.IsInf() {
		return true
	}

	validateModulo(elem.DataX.M, elem.DataY.M)

	calcY2 := elem.ElemParams.calcYSquared(elem.DataX)
	calcY2Check := elem.DataY.Square()

	return calcY2.IsValEqual(calcY2Check)
}

func (elem *CurveElement) isEqual(cmpElem *CurveElement) bool {
	if !elem.DataX.IsValEqual(cmpElem.DataX) {
		return false
	}
	return elem.DataY.IsValEqual(cmpElem.DataY)
}

func (elem *CurveElement) CopyPow() IPowElement {
	theCopy := elem.dup()
	theCopy.freeze()
	return theCopy
}

func (elem *CurveElement) dup() *CurveElement {
	newElem := new(CurveElement)
	newElem.ElemParams = elem.ElemParams
	newElem.DataX = elem.DataX.Copy()
	newElem.DataY = elem.DataY.Copy()
	return newElem
}

func (elem *CurveElement) MakeOnePow() IPowElement {
	return &CurveElement{elem.ElemParams, PointLike{nil, nil}}
}

func (elem *CurveElement) MulPoint(elemIn *CurveElement) *CurveElement {
	res := elem.mul(elemIn)
	return res
}

func (elem *CurveElement) MulPow(elemIn IPowElement) IPowElement {
	res := elem.mul(elemIn.(*CurveElement))
	return res
}

func (elem *CurveElement) set(in *CurveElement) {
	elem.DataX = in.DataX
	elem.DataY = in.DataY
}

func (elem *CurveElement) twiceInternal() *CurveElement {

	if !elem.frozen() {
		panic("elem input must be Frozen")
	}

	// We have P1 = P2 so the tangent line T at P1 ha slope
	// lambda = (3x^2 + A) / 2y
	lambdaNumer := elem.DataX.Square().Mul(MI_THREE).Add(elem.ElemParams.A.ModInt)
	lambdaDenom := elem.DataY.Add(elem.DataY).Invert()
	lambda := lambdaNumer.Mul(lambdaDenom)
	lambda.Freeze()

	// x3 = lambda^2 - 2x
	x3 := lambda.Square().Sub(elem.DataX.Add(elem.DataX))

	// y3 = (x - x3) lambda - y
	y3 := elem.DataX.Sub(x3).Mul(lambda).Sub(elem.DataY)

	x3.Freeze()
	y3.Freeze()
	return &CurveElement{elem.ElemParams, PointLike{x3, y3}}
}

func (elem *CurveElement) mul(elemIn *CurveElement) *CurveElement {

	if !elemIn.frozen() {
		panic("elemIn param must be Frozen")
	}

	if elem.IsInf() {
		return elemIn
	}

	if elemIn.IsInf() {
		return elem
	}

	if elem.DataX.IsValEqual(elemIn.DataX) {
		if elem.DataY.IsValEqual(elemIn.DataY) {
			if elem.DataY.IsValEqual(MI_ZERO) {
				return &CurveElement{elem.ElemParams, PointLike{nil, nil}}
			} else {
				return elem.twiceInternal()
			}
		}
		return &CurveElement{elem.ElemParams, PointLike{nil, nil}}
	}

	// P1 != P2, so the slope of the line L through P1 and P2 is
	// lambda = (y2-y1)/(x2-x1)
	lambdaNumer := elemIn.DataY.Sub(elem.DataY)
	lambdaDenom := elemIn.DataX.Sub(elem.DataX)
	lambda := lambdaNumer.Mul(lambdaDenom.Invert())
	lambda.Freeze()

	// x3 = lambda^2 - x1 - x2
	x3 := lambda.Square().Sub(elem.DataX).Sub(elemIn.DataX)

	// y3 = (x1-x3) lambda - y1
	y3 := elem.DataX.Sub(x3).Mul(lambda).Sub(elem.DataY)

	x3.Freeze()
	y3.Freeze()
	return &CurveElement{elem.ElemParams, PointLike{x3, y3}}
}
