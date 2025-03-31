package gencrypt

import (
	"github.com/qiafan666/gotato/commons/gcommon"
	"testing"
)

func TestAkSk(t *testing.T) {

	ak, sk := GenerateAkSk()
	t.Log("ak:", ak)
	t.Log("sk:", sk)

	testMap := map[string]any{
		"ak":    ak,
		"test":  1,
		"sk":    sk,
		"slice": []int{1, 2, 3},
		"map":   map[string]int{"a": 1, "b": 2},
		"bool":  true,
		"obj":   struct{ Name string }{"test"},
		"float": 3.14,
	}

	testString := gcommon.MapSortUrl(testMap)
	t.Log("testString:", testString)

	signature := GenerateSignature(sk, testString)
	t.Log("signature:", signature)

	if !VerifySignature(sk, testString, signature) {
		t.Error("VerifySignature failed")
	} else {
		t.Log("VerifySignature success")
	}
}
