package gcommon

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/jinzhu/copier"
	"github.com/qiafan666/gotato/commons/gcast"
	"golang.org/x/exp/constraints"
	"gorm.io/gorm"
	"math"
	"reflect"
	"regexp"
	"strings"
)

// GetRequestId 获取request_id
func GetRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestId, ok := ctx.Value("request_id").(string); ok {
		return requestId
	} else {
		return ""
	}
}

func GetRequestIdFormat(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestId, ok := ctx.Value("request_id").(string); ok {
		return fmt.Sprintf("[request_id:%s] ", requestId)
	} else {
		return ""
	}
}

func SetRequestId(requestId string) context.Context {
	return context.WithValue(context.Background(), "request_id", requestId)
}

func SetRequestIdWithCtx(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, "request_id", requestId)
}

// RetryFunction 重试函数
func RetryFunction(c func() bool, times int) bool {
	for i := times + 1; i > 0; i-- {
		if c() == true {
			return true
		}
	}
	return false
}

// VersionCompare 语义化的版本比较，支持：>, >=, =, !=, <, <=, | (or), & (and).
// 参数 `rangeVer` 示例：1.0.0, =1.0.0, >2.0.0, >=1.0.0&<2.0.0, <2.0.0|>3.0.0, !=4.0.4
func VersionCompare(rangeVer, curVer string) (bool, error) {
	semVer, err := version.NewVersion(curVer)
	if err != nil {
		return false, err
	}

	orVers := strings.Split(rangeVer, "|")
	for _, ver := range orVers {
		andVers := strings.Split(ver, "&")
		constraints, err := version.NewConstraint(strings.Join(andVers, ","))
		if err != nil {
			return false, err
		}
		if constraints.Check(semVer) {
			return true, nil
		}
	}
	return false, nil
}

// StructToMap 筛选出非nil的字段，转换成map,跳过指定字段，json标签为空的字段，json标签为数据库字段
// JumpString 跳过指定字段 不解析第二层struct
func StructToMap(inputStruct interface{}, JumpString ...string) map[string]interface{} {

	structValue := reflect.ValueOf(inputStruct)
	structType := structValue.Type()

	resultMap := make(map[string]interface{})
	if structType.Kind() != reflect.Struct {
		return resultMap
	}

	for i := 0; i < structValue.NumField(); i++ {
		fieldValue := structValue.Field(i)
		fieldName := structType.Field(i).Name
		if len(JumpString) > 0 {
			if SliceContain(JumpString, fieldName) {
				continue
			}
		}

		if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
			continue // 跳过 nil 值的字段
		}

		if len(structType.Field(i).Tag) == 0 || len(structType.Field(i).Tag.Get("json")) == 0 || structType.Field(i).Tag.Get("json") == "-" {
			continue
		}

		resultMap[structType.Field(i).Tag.Get("json")] = fieldValue.Interface()
	}
	return resultMap
}

// Paginate 分页
func Paginate(pageNum interface{}, pageSize interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if gcast.ToInt(pageNum) <= 0 {
			pageNum = 1
		}
		switch {
		case gcast.ToInt(pageSize) > 1000:
			pageSize = 100
		case gcast.ToInt(pageSize) <= 0:
			pageSize = 1
		}
		offset := (gcast.ToInt(pageNum) - 1) * gcast.ToInt(pageSize)
		return db.Offset(offset).Limit(gcast.ToInt(pageSize))
	}
}

// CopyStructFields 复制结构体字段
func CopyStructFields(from any, to any) (err error) {
	return copier.Copy(to, from)
}

func Kvs(kv ...any) string {
	return Kv2Str("", kv...)
}

func Kv2Str(msg string, kv ...any) string {
	if len(kv) == 0 {
		return msg
	} else {
		var buf bytes.Buffer
		buf.WriteString(msg)

		for i := 0; i < len(kv); i += 2 {
			if buf.Len() > 0 {
				buf.WriteString(", ")
			}

			key := fmt.Sprintf("%v", kv[i])
			buf.WriteString(key)
			buf.WriteString("=")

			if i+1 < len(kv) {
				value := fmt.Sprintf("%v", kv[i+1])
				buf.WriteString(value)
			} else {
				buf.WriteString("MISSING")
			}
		}
		return buf.String()
	}
}

// HideStr 隐藏字符串,包括邮箱和手机号,身份证号保留前三后四
func HideStr(str string) (result string) {
	if str == "" {
		return "***"
	}

	email, b := HideEmail(str)
	if b {
		return email
	}

	phone, b := HidePhone(str)
	if b {
		return phone
	}

	nameRune := []rune(str)
	lens := len(nameRune)
	if lens <= 1 {
		result = "***"
	} else if lens == 2 {
		result = string(nameRune[:1]) + "*"
	} else if lens == 3 {
		result = string(nameRune[:1]) + "*" + string(nameRune[2:3])
	} else if lens == 4 {
		result = string(nameRune[:1]) + "**" + string(nameRune[lens-1:lens])
	} else if lens > 4 && lens <= 7 {
		result = string(nameRune[:2]) + "***" + string(nameRune[lens-2:lens])
	} else if lens > 7 {
		result = string(nameRune[:3]) + "***" + string(nameRune[lens-4:lens])
	}
	return result
}
func SubStr(str string, start int, end int) string {
	rs := []rune(str)
	return string(rs[start:end])
}

// HideEmail 隐藏邮箱
func HideEmail(email string) (string, bool) {
	if strings.Contains(email, "@") {
		// 邮箱
		res := strings.Split(email, "@")
		if len(res[0]) < 3 {
			resString := "***"
			email = resString + "@" + res[1]
		} else {
			res2 := SubStr(email, 0, 3)
			resString := res2 + "***"
			email = resString + "@" + res[1]
		}
		return email, true
	}
	return "", false
}

// HidePhone 隐藏手机号
func HidePhone(phone string) (string, bool) {
	reg := `^1[0-9]\d{9}$`
	rgx := regexp.MustCompile(reg)
	mobileMatch := rgx.MatchString(phone)
	if mobileMatch {
		// 手机号
		return SubStr(phone, 0, 3) + "****" + SubStr(phone, 7, 11), true
	}
	return "", false
}

// MultiPointDistance 多个坐标点之间的距离
// 例如：x1, x2, y1, y2
func MultiPointDistance(p ...float64) float64 {
	if len(p)%2 != 0 {
		return 0
	}
	var sum float64
	var i = 0
	for {
		if i >= len(p) {
			break
		}
		sum += math.Pow(p[i]-p[i+1], 2)
		i += 2
	}
	return math.Sqrt(sum)
}

// Max 最大值
func Max[T constraints.Ordered](nums ...T) T {
	if len(nums) == 0 {
		panic("Max: 至少需要一个数")
	}
	maxVal := nums[0]
	for _, num := range nums[1:] {
		if num > maxVal {
			maxVal = num
		}
	}
	return maxVal
}

// Min 最小值
func Min[T constraints.Ordered](nums ...T) T {
	if len(nums) == 0 {
		panic("Min: 至少需要一个数")
	}
	minVal := nums[0]
	for _, num := range nums[1:] {
		if num < minVal {
			minVal = num
		}
	}
	return minVal
}
