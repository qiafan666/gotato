package gcommon

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcast"
	uuid "github.com/satori/go.uuid"
	"hash/fnv"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

// StringToSha256 字符串转SHA256
func StringToSha256(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte("hello world\n")))
}

const Lower = "abcdefghijklmnopqrstuvwxyz"
const Upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const Number = "0123456789"
const Symbol = "!@#$%^&*()_+-=[]{}|;':\",./<>?"
const mixString = Lower + Upper + Number
const lowerUpper = Lower + Upper

// RandLower 生成随机小写
func RandLower(stringSize int) string {
	return randomStr(Lower, stringSize)
}

// RandUpper 生成随机大写字母
func RandUpper(stringSize int) string {
	return randomStr(Upper, stringSize)
}

// RandNum 生成随机数字
func RandNum(stringSize int) string {
	return randomStr(Number, stringSize)
}

// RandSymbol 生成随机符号
func RandSymbol(stringSize int) string {
	return randomStr(Symbol, stringSize)
}

// RandLowerUpper 生成大小写英文混合随机字符串
func RandLowerUpper(stringSize int) string {
	return randomStr(lowerUpper, stringSize)
}

// RandStr 生成随机字符串
func RandStr(stringSize int) string {
	return randomStr(mixString, stringSize)
}

// RandCusStr 生成自定义字符串
func RandCusStr(src string, length int) string {
	return randomStr(src, length)
}

func randomStr(str string, length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		for j := range b {
			b[j] = str[r.Intn(len(str))]
		}
	}
	return string(b)
}

type NumberType interface {
	int | int32 | int64 | float32 | float64
}

func RangeNum[T NumberType](min, max T) T {
	if min > max {
		panic("min must be less than or equal to max")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch any(min).(type) {
	case float32, float64:
		return T(r.Float64()*float64(max-min) + float64(min))
	default:
		return T(r.Intn(int(max-min+1)) + int(min))
	}
}

// StrCheck 检查输入的字符串是否有空值
func StrCheck(input ...string) []int {
	var nullIndices []int

	for i, data := range input {
		if len(data) == 0 {
			nullIndices = append(nullIndices, i)
		}
	}
	if len(nullIndices) > 0 {
		return nullIndices
	}
	return nullIndices
}

// StrToBytes 原地转换
func StrToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Str2BytesNoCopy
// 无拷贝 string 转 []byte
func Str2BytesNoCopy(s string) []byte {
	tmp := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{tmp[0], tmp[1], tmp[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// UnderscoreName 驼峰式写法转为下划线写法
func UnderscoreName(name string) string {
	// 使用 strings.Builder 代替自定义的 NewBuffer
	var buffer strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.WriteRune('_') // 使用 WriteRune 追加下划线字符
			}
			buffer.WriteRune(unicode.ToLower(r)) // 将大写字母转为小写
		} else {
			buffer.WriteRune(r) // 保持非大写字母原样追加
		}
	}

	return buffer.String()
}

// CamelName 下划线写法转为驼峰式写法
func CamelName(name string, firstUpper bool) string {
	var buffer strings.Builder
	skipNext := false
	for i, r := range name {
		if skipNext {
			skipNext = false
			continue
		}
		if r == '_' {
			if i+1 < len(name) {
				buffer.WriteRune(unicode.ToUpper(rune(name[i+1])))
				skipNext = true // 跳过下划线后的字符
			}
		} else {
			// 对于首字母的处理
			if i == 0 && firstUpper {
				buffer.WriteRune(unicode.ToUpper(r))
			} else {
				buffer.WriteRune(r)
			}
		}
	}
	return buffer.String()
}

// Byte2Uint32 字节数组转uint32
func Byte2Uint32(b []byte) uint32 {
	seed := uint32(131)
	hash := uint32(0)

	for _, v := range b {
		hash = hash*seed + uint32(v)
	}
	return hash
}

// Str2Uint32 string 转 uint32
func Str2Uint32(s string) uint32 {
	b := bytes.NewBufferString(s).Bytes()
	return Byte2Uint32(b)
}

// CountRune 计算包含中文的字符串长度，但是一个中文算3个长度
func CountRune(str string) int {
	length := 0
	for _, runeValue := range str {
		length += utf8.RuneLen(runeValue)
	}
	return length
}

// ParseChinese 提取字符串中的中文
func ParseChinese(str string) string {
	b := strings.Builder{}
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			b.WriteRune(v)
		}
	}
	return b.String()
}

// StrJoin 字符串连接
func StrJoin(sep string, str ...any) string {
	sb := strings.Builder{}

	// 预先分配足够的内存，以减少内存分配和复制的开销
	total := len(str) - 1
	for _, s := range str {
		total += len(gcast.ToString(s))
	}
	sb.Grow(total)

	for i, s := range str {
		sb.WriteString(gcast.ToString(s))
		if i < len(str)-1 {
			sb.WriteString(sep)
		}
	}
	return sb.String()
}

// StrParse 字符串解析
func StrParse(str string, sep string) []string {
	return strings.Split(str, sep)
}

// BuildString 拼接字符串
func BuildString(members ...any) string {
	sb := strings.Builder{}
	for _, m := range members {
		sb.WriteString(gcast.ToString(m))
	}
	return sb.String()
}

// ContainsEmoji 检查字符串中是否包含表情符号
func ContainsEmoji(s string) bool {
	emojiPattern := `[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F700}-\x{1F77F}]|[\x{1F780}-\x{1F7FF}]|[\x{1F800}-\x{1F8FF}]|[\x{1F900}-\x{1F9FF}]|[\x{1FA00}-\x{1FA6F}]|[\x{1FA70}-\x{1FAFF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`
	re := regexp.MustCompile(emojiPattern)
	return re.MatchString(s)
}

func HashString(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
