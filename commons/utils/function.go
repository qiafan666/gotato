package utils

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/qiafan666/gotato/commons"
	"github.com/qiafan666/gotato/commons/log"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"time"
)

func GenerateUUID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

func StringToMd5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func StringsToString(stringArray []string) string {
	if len(stringArray) <= 0 {
		return ""
	}
	return strings.Join(stringArray, ",")
}

func StringToStrings(param string) []string {
	if len(param) <= 0 {
		return []string{}
	}
	return strings.Split(param, ",")
}

func StringToSha256(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte("hello world\n")))
}
func RetryFunction(c func() bool, times int) bool {
	for i := times + 1; i > 0; i-- {
		if c() == true {
			return true
		}
	}
	return false
}

func ValidateAndBindParameters(entity interface{}, ctx iris.Context, info string) (commons.ResponseCode, string) {
	if err := ctx.UnmarshalBody(entity, iris.UnmarshalerFunc(json.Unmarshal)); err != nil {
		log.Slog.ErrorF(ctx.Values().Get("ctx").(context.Context), "%s error %s", info, err.Error())
		return commons.ParameterError, err.Error()
	}
	if err := Validate(entity); err != nil {
		log.Slog.ErrorF(ctx.Values().Get("ctx").(context.Context), "%s error %s", info, err.Error())
		return commons.ValidateError, err.Error()
	}

	return commons.OK, ""
}

func ValidateAndBindCtxParameters(entity interface{}, ctx iris.Context, info string) (commons.ResponseCode, string) {
	err := json.Unmarshal(ctx.Values().Get(commons.CtxValueParameter).([]byte), entity)
	if err != nil {
		log.Slog.ErrorF(ctx.Values().Get("ctx").(context.Context), "%s error %s", info, err.Error())
		return commons.ParameterError, err.Error()
	}
	if err := Validate(entity); err != nil {
		log.Slog.ErrorF(ctx.Values().Get("ctx").(context.Context), "%s error %s", info, err.Error())
		return commons.ValidateError, err.Error()
	}

	return commons.OK, ""
}

// DifferenceInt64 找出两个数组不存在的元素
func DifferenceInt64(a, b []int64) []int64 {
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
	var diff []int64
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			diff = append(diff, a[i])
			i++
		} else if a[i] > b[j] {
			diff = append(diff, b[j])
			j++
		} else {
			i++
			j++
		}
	}
	for ; i < len(a); i++ {
		diff = append(diff, a[i])
	}
	for ; j < len(b); j++ {
		diff = append(diff, b[j])
	}
	return diff
}

// CommonInt64 找出两个数组相同的元素
func CommonInt64(a, b []int64) []int64 {
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
	var common []int64
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if a[i] > b[j] {
			j++
		} else {
			common = append(common, a[i])
			i++
			j++
		}
	}
	return common
}

// DifferenceStrings 找出两个数组不存在的元素
func DifferenceStrings(a, b []string) []string {
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
	var diff []string
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			diff = append(diff, a[i])
			i++
		case a[i] > b[j]:
			diff = append(diff, b[j])
			j++
		default:
			i++
			j++
		}
	}
	for ; i < len(a); i++ {
		diff = append(diff, a[i])
	}
	for ; j < len(b); j++ {
		diff = append(diff, b[j])
	}
	return diff
}

// CommonStrings 找出两个数组相同的元素
func CommonStrings(a, b []string) []string {
	sort.Strings(a)
	sort.Strings(b)
	var common []string
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if a[i] > b[j] {
			j++
		} else {
			common = append(common, a[i])
			i++
			j++
		}
	}
	return common
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(stringSize int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
	}
	return strings.ToUpper(string(b))
}

// 数组中是否包含某个元素
func Contains[T string | int64](ts []T, t T) bool {
	for _, v := range ts {
		if reflect.DeepEqual(v, t) {
			return true
		}
	}

	return false
}

// 使用泛型函数来删除切片中的某个元素
func Remove[T any](ts []T, t T) []T {
	for i, v := range ts {
		if reflect.DeepEqual(v, t) {
			return append(ts[:i], ts[i+1:]...)
		}
	}
	return ts // 如果未找到匹配的元素，则返回原始切片
}

// 合并两个切片，去除重复元素
func MergeString(a []string, b []string) []string {
	// 创建一个map用于存储所有元素的唯一值
	uniqueElements := make(map[string]bool)

	// 遍历数组a，将其中的元素添加到map中
	for _, element := range a {
		uniqueElements[element] = true
	}

	// 遍历数组b，将其中的元素添加到map中
	for _, element := range b {
		uniqueElements[element] = true
	}

	// 将map中的唯一值提取到一个新的切片中
	merged := []string{}
	for element := range uniqueElements {
		merged = append(merged, element)
	}

	return merged
}

// 合并两个切片，去除重复元素
func MergeInt64(a []int64, b []int64) []int64 {
	// 创建一个map用于存储所有元素的唯一值
	uniqueElements := make(map[int64]bool)

	// 遍历数组a，将其中的元素添加到map中
	for _, element := range a {
		uniqueElements[element] = true
	}

	// 遍历数组b，将其中的元素添加到map中
	for _, element := range b {
		uniqueElements[element] = true
	}

	// 将map中的唯一值提取到一个新的切片中
	merged := []int64{}
	for element := range uniqueElements {
		merged = append(merged, element)
	}

	return merged
}
