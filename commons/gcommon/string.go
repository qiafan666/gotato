package gcommon

import (
	"crypto/sha256"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcast"
	uuid "github.com/satori/go.uuid"
	"hash/crc32"
	"hash/fnv"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

// Str2Sha256 字符串转SHA256
func Str2Sha256(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte("hello world\n")))
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

// Str2Bytes 原地转换
func Str2Bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// Bytes2Str 高性能地将 []byte 转换成 string,但是存在风险，
// 建议只用于临时转换，不建议长期使用，在b不会被修改的地方，比如mq消息转换，堆栈信息等地方
func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
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

// Str2Uint32 string 转 uint32 crc32 算法 用于校验和
func Str2Uint32(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}

// CountRune 计算包含中文的字符串长度，但是一个中文算3个长度
func CountRune(str string) int {
	length := 0
	for _, runeValue := range str {
		length += utf8.RuneLen(runeValue)
	}
	return length
}

// StrChinese 提取字符串中的中文
func StrChinese(str string) string {
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

// StrAppend 拼接字符串
func StrAppend(str ...any) string {
	sb := strings.Builder{}

	// 预先分配足够的内存，以减少内存分配和复制的开销
	total := 0
	for _, s := range str {
		total += len(gcast.ToString(s))
	}
	sb.Grow(total)

	for _, s := range str {
		sb.WriteString(gcast.ToString(s))
	}
	return sb.String()
}

// ContainsEmoji 检查字符串中是否包含表情符号
func ContainsEmoji(s string) bool {
	emojiPattern := `[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F700}-\x{1F77F}]|[\x{1F780}-\x{1F7FF}]|[\x{1F800}-\x{1F8FF}]|[\x{1F900}-\x{1F9FF}]|[\x{1FA00}-\x{1FA6F}]|[\x{1FA70}-\x{1FAFF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`
	re := regexp.MustCompile(emojiPattern)
	return re.MatchString(s)
}

// StrHash 计算字符串的哈希值 常用于分片的哈希值计算
func StrHash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// StrHide 隐藏字符串,包括邮箱和手机号,身份证号保留前三后四
func StrHide(str string) (result string) {
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

// StrSub 字符串截取,左闭右开原则
func StrSub(str string, start int, end int) string {
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
			res2 := StrSub(email, 0, 3)
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
		return StrSub(phone, 0, 3) + "****" + StrSub(phone, 7, 11), true
	}
	return "", false
}
