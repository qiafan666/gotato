package gcommon

import (
	"context"
	"crypto/md5"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/kataras/iris/v12"
	"github.com/qiafan666/gotato/commons"
	"github.com/qiafan666/gotato/commons/log"
	"github.com/qiafan666/gotato/commons/utils"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"io"
	mathRand "math/rand"
	"reflect"
	"strings"
	"time"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

// StringToMd5 字符串转MD5
func StringToMd5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

// StringsToString 字符串数组转字符串
func StringsToString(stringArray []string) string {
	if len(stringArray) <= 0 {
		return ""
	}
	return strings.Join(stringArray, ",")
}

// StringToStrings 字符串转字符串数组;,逗号分隔
func StringToStrings(param string) []string {
	if len(param) <= 0 {
		return []string{}
	}
	return strings.Split(param, ",")
}

// StringToSha256 字符串转SHA256
func StringToSha256(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte("hello world\n")))
}

// Nonce 生成随机串(size应为偶数)
func Nonce(size uint8) string {
	nonce := make([]byte, size/2)
	io.ReadFull(cryptoRand.Reader, nonce)
	return hex.EncodeToString(nonce)
}

const randomLowerUpperNumberMixed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const randomLower = "abcdefghijklmnopqrstuvwxyz"
const randomUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const randomNumber = "0123456789"
const randomLowerUpper = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandomLowerUpperNumberMixed 生成随机字符串
func RandomLowerUpperNumberMixed(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomLowerUpperNumberMixed[mathRand.Intn(len(randomLowerUpperNumberMixed))]
		}
	}
	return string(b)
}

// RandomLower 生成随机小写
func RandomLower(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomLower[mathRand.Intn(len(randomLower))]
		}
	}
	return string(b)
}

// RandomUpper 生成随机大写字母
func RandomUpper(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomUpper[mathRand.Intn(len(randomUpper))]
		}
	}
	return string(b)
}

// RandomNumber 生成随机数字
func RandomNumber(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomNumber[mathRand.Intn(len(randomNumber))]
		}
	}
	return string(b)
}

// RandomLowerUpperMixed 生成大小写英文混合随机字符串
func RandomLowerUpperMixed(stringSize int) string {

	mathRand.Seed(time.Now().UnixNano())
	b := make([]byte, stringSize)
	for i := 0; i < stringSize; i++ {
		for i := range b {
			b[i] = randomLowerUpper[mathRand.Intn(len(randomLowerUpper))]
		}
	}
	return string(b)
}

// DataCheck 检查输入的字符串是否有空值
func DataCheck(input ...string) []int {
	var nullIndices []int

	for i, data := range input {
		if len(data) == 0 {
			nullIndices = append(nullIndices, i)
		}
	}

	return nullIndices
}

// -------------------------- 集合相关函数 --------------------------

// SliceContain 返回指定元素是否在集合中
func SliceContain[T ~[]E, E comparable](list T, elem E) bool {
	if len(list) == 0 {
		return false
	}
	for _, v := range list {
		if v == elem {
			return true
		}
	}
	return false
}

// SliceUniq 集合去重
func SliceUniq[T ~[]E, E comparable](list T) T {
	if len(list) == 0 {
		return list
	}

	ret := make(T, 0, len(list))
	m := make(map[E]struct{}, len(list))
	for _, v := range list {
		if _, ok := m[v]; !ok {
			ret = append(ret, v)
			m[v] = struct{}{}
		}
	}
	return ret
}

// SliceDiff 返回两个集合之间的差异
func SliceDiff[T ~[]E, E comparable](list1 T, list2 T) (ret1 T, ret2 T) {
	m1 := map[E]struct{}{}
	m2 := map[E]struct{}{}
	for _, v := range list1 {
		m1[v] = struct{}{}
	}
	for _, v := range list2 {
		m2[v] = struct{}{}
	}

	ret1 = make(T, 0)
	ret2 = make(T, 0)
	for _, v := range list1 {
		if _, ok := m2[v]; !ok {
			ret1 = append(ret1, v)
		}
	}
	for _, v := range list2 {
		if _, ok := m1[v]; !ok {
			ret2 = append(ret2, v)
		}
	}
	return ret1, ret2
}

// SliceWithout 返回不包括所有给定值的切片
func SliceWithout[T ~[]E, E comparable](list T, exclude ...E) T {
	if len(list) == 0 {
		return list
	}

	m := make(map[E]struct{}, len(exclude))
	for _, v := range exclude {
		m[v] = struct{}{}
	}

	ret := make(T, 0, len(list))
	for _, v := range list {
		if _, ok := m[v]; !ok {
			ret = append(ret, v)
		}
	}
	return ret
}

// SliceIntersect 返回两个集合的交集
func SliceIntersect[T ~[]E, E comparable](list1 T, list2 T) T {
	m := make(map[E]struct{})
	for _, v := range list1 {
		m[v] = struct{}{}
	}

	ret := make(T, 0)
	for _, v := range list2 {
		if _, ok := m[v]; ok {
			ret = append(ret, v)
		}
	}
	return ret
}

// SliceUnion 返回两个集合的并集
func SliceUnion[T ~[]E, E comparable](lists ...T) T {
	ret := make(T, 0)
	m := make(map[E]struct{})
	for _, list := range lists {
		for _, v := range list {
			if _, ok := m[v]; !ok {
				ret = append(ret, v)
				m[v] = struct{}{}
			}
		}
	}
	return ret
}

// SliceRand 返回一个指定随机挑选个数的切片
// 若 n == -1 or n >= len(list)，则返回打乱的切片
func SliceRand[T ~[]E, E any](list T, n int) T {
	if n == 0 || n < -1 {
		return nil
	}

	count := len(list)
	ret := make(T, count)
	copy(ret, list)
	mathRand.Shuffle(count, func(i, j int) {
		ret[i], ret[j] = ret[j], ret[i]
	})
	if n == -1 || n >= count {
		return ret
	}
	return ret[:n]
}

// SlicePinTop 置顶集合中的一个元素
func SlicePinTop[T any](list []T, index int) {
	if index <= 0 || index >= len(list) {
		return
	}
	for i := index; i > 0; i-- {
		list[i], list[i-1] = list[i-1], list[i]
	}
}

// SlicePinTopF 置顶集合中满足条件的一个元素
func SlicePinTopF[T any](list []T, fn func(v T) bool) {
	index := 0
	for i, v := range list {
		if fn(v) {
			index = i
			break
		}
	}
	for i := index; i > 0; i-- {
		list[i], list[i-1] = list[i-1], list[i]
	}
}

// SliceAppendUnique 数组中是否包含某个元素,没有就追加,有就返回原数组
func SliceAppendUnique[T any](ts []T, t T) []T {
	for _, v := range ts {
		if reflect.DeepEqual(v, t) {
			return ts
		}
	}
	ts = append(ts, t)
	return ts
}

// SliceRemove 使用泛型函数来删除切片中的某个元素
func SliceRemove[T any](ts []T, t T) []T {
	for i, v := range ts {
		if reflect.DeepEqual(v, t) {
			return append(ts[:i], ts[i+1:]...)
		}
	}
	return ts // 如果未找到匹配的元素，则返回原始切片
}

// ----------------------- 其他函数 -----------------------

// RetryFunction 重试函数
func RetryFunction(c func() bool, times int) bool {
	for i := times + 1; i > 0; i-- {
		if c() == true {
			return true
		}
	}
	return false
}

// ValidateAndBindCtxParameters iris v2版本参数验证
func ValidateAndBindCtxParameters(entity interface{}, ctx iris.Context, info string) (commons.ResponseCode, string) {
	err := json.Unmarshal(ctx.Values().Get(commons.CtxValueParameter).([]byte), entity)
	if err != nil {
		log.Slog.ErrorF(ctx.Values().Get("ctx").(context.Context), "%s error %s", info, err.Error())
		return commons.ParameterError, err.Error()
	}
	if err := utils.Validate(entity); err != nil {
		log.Slog.ErrorF(ctx.Values().Get("ctx").(context.Context), "%s error %s", info, err.Error())
		return commons.ValidateError, err.Error()
	}

	return commons.OK, ""
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

// StructToStringMapWithNilFilter 筛选出非nil的字段，转换成map,用于更新数据库,跳过指定字段，json标签为空的字段，json标签为数据库字段
func StructToStringMapWithNilFilter(inputStruct interface{}, table string, JumpString ...string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	resultMap[commons.Table] = table

	structValue := reflect.ValueOf(inputStruct)
	structType := structValue.Type()

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
func Paginate(pageNum int, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pageNum == 0 {
			pageNum = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 8
		}
		offset := (pageNum - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
