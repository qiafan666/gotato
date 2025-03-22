package gencrypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gcommon"
	"sort"
	"strings"
)

// GenerateAkSk 生成ak sk
func GenerateAkSk() (ak string, sk string) {
	ak = gcommon.GenerateUUID()
	sk = gcommon.GenerateUUID()
	return
}

// GenerateSignature  hmac sha256 签名
// 调用方使用sk签名msg
// 调用方将ak，msg，签名一起发送给服务方
func GenerateSignature(sk, msg string) string {
	h := hmac.New(sha256.New, []byte(sk))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature 验证签名
// 服务方验证调用方的签名是否正确
// 服务方通过调用方的ak获取sk，然后使用sk验证签名是否正确
func VerifySignature(sk, msg, signMsg string) bool {
	expectedSignature := GenerateSignature(sk, msg)
	return hmac.Equal([]byte(expectedSignature), []byte(signMsg))
}

// MapSort 根据字符顺序排序map并拼接成字符串
func MapSort(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var queryString strings.Builder
	for i, key := range keys {
		queryString.WriteString(fmt.Sprintf("%s=%s", key, gcast.ToString(m[key])))
		if i < len(keys)-1 {
			queryString.WriteString("&")
		}
	}
	return queryString.String()
}
