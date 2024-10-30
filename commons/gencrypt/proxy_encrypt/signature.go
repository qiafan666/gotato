package proxy_encrypt

//func Sign2(cxt *Context, msg []byte, seckey *UmbralFieldElement) (*UmbralFieldElement, *UmbralFieldElement) {
//	h:=GenPrivateKeyFromMsg(cxt,msg)
//	r := &UmbralFieldElement{*cxt.targetField.NewElement(ONE)}//GenPrivateKey(cxt)
//	//r := &UmbralFieldElement{*cxt.targetField.NewElement(cxt.targetField.FieldOrder)}//GenPrivateKey(cxt)
//	rG := r.GetPublicKey(cxt)
//	x:=rG.DataX
//	_s:=seckey.Mul(x).Add(h.ModInt).Mul(r.Invert())
//	//_s1:=(seckey.Mul(x).Mul(r.Invert())).Add(h.Mul(r.Invert()))
//	//fmt.Println(_s,_s1)
//	e := cxt.targetField.NewElement(_s.GetValue())
//	s:=&UmbralFieldElement{*e}
//	//fmt.Println(_s,s)
//	return r,s
//}
//
//func Verify2(cxt *Context, r *UmbralFieldElement, s *UmbralFieldElement, msg []byte, pubKey *UmbralCurveElement) bool{
//	h:=GenPrivateKeyFromMsg(cxt,msg)
//	rG := r.GetPublicKey(cxt)
//	h_s:=h.Mul(s.Invert()).GetValue()
//	x_s:=rG.DataX.Mul(s.Invert()).GetValue()
//	h_sG:=cxt.curveField.GetGen().MulScalar(h_s)
//	x_sPub:=pubKey.MulScalar(x_s)
//	fmt.Println(h_sG.Add(x_sPub))
//	fmt.Println(rG)
//	//fmt.Println(h_sG.Add(x_sPub).PointLike.DataX,rG.PointLike.DataX)
//	return rG.PointLike == h_sG.Add(x_sPub).PointLike
//}

func Sign(cxt *Context, msg []byte, seckey *UmbralFieldElement) (*UmbralFieldElement, *UmbralFieldElement) {
	h := GenPrivateKeyFromMsg(cxt, msg)
	k := &UmbralFieldElement{*cxt.targetField.NewElement(ONE)} //GenPrivateKey(cxt)
	//r := &UmbralFieldElement{*cxt.targetField.NewElement(cxt.targetField.FieldOrder)}//GenPrivateKey(cxt)
	kG := k.GetPublicKey(cxt)
	_r := kG.DataX
	e := cxt.targetField.NewElement(_r.GetValue())
	r := &UmbralFieldElement{*e}

	_s := seckey.Mul(r.ModInt).Add(h.ModInt).Mul(k.Invert())
	//_s1:=(seckey.Mul(x).Mul(r.Invert())).Add(h.Mul(r.Invert()))
	//fmt.Println(_s,_s1)
	e = cxt.targetField.NewElement(_s.GetValue())
	s := &UmbralFieldElement{*e}
	//fmt.Println(_s,s)
	return r, s
}

func Verify(cxt *Context, r *UmbralFieldElement, s *UmbralFieldElement, msg []byte, pubKey *UmbralCurveElement) bool {
	h := GenPrivateKeyFromMsg(cxt, msg)
	s_1 := s.Invert()
	h_s1 := h.Mul(s_1).GetValue()
	r_s1 := r.Mul(s_1).GetValue()
	P_1 := cxt.curveField.GetGen().MulScalar(h_s1)
	P_2 := pubKey.MulScalar(r_s1)
	R2 := P_1.Add(P_2)
	//fmt.Println(r.GetValue(), R2.DataX.GetValue())
	return r.GetValue().Cmp(R2.DataX.GetValue()) == 0
}
