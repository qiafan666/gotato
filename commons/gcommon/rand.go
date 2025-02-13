package gcommon

import (
	"math/rand"
	"time"
)

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

const (
	DefaultRandWeight = 10000
)

type RandInfo struct {
	Weight int32
	Value  any
}

// RandOnce 在randList中按权重随机一个value
func RandOnce(randList []RandInfo) any {
	if len(randList) == 0 {
		return nil
	}
	if len(randList) == 1 {
		return randList[0].Value
	}
	var total int32
	for _, info := range randList {
		total += info.Weight
	}
	if total <= 0 {
		return nil
	}
	randNum := rand.Int31n(total)
	for _, info := range randList {
		if info.Weight >= randNum {
			return info.Value
		}
		randNum -= info.Weight
	}
	return nil
}

// RandNumNoPutBack 在sample（key:valueID, value:weight） 不放回抽取num个样本
func RandNumNoPutBack(num int32, sample map[int32]int32) []int32 {
	values := make([]int32, 0, num)
	if len(sample) < int(num) || num <= 0 {
		return nil
	}
	var total int32
	for _, weight := range sample {
		total += weight
	}
	if total <= 0 {
		return nil
	}
	for num > 0 {
		randNum := rand.Int31n(total)
		for value, weight := range sample {
			if weight >= randNum {
				values = append(values, value)
				delete(sample, value)
				break
			}
			randNum -= weight
		}
		num--
	}
	return values
}

// RandRedPacket 随机红包算法
func RandRedPacket(num int32, totalCount int64) []int64 {
	// 确保红包数不大于总金额
	if num <= 0 || totalCount < int64(num) {
		return nil // 或者返回错误
	}

	envelopes := make([]int64, num)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	remainingCount := totalCount

	// 先给每个红包分配1元，确保每个红包至少有1元
	for i := 0; i < int(num); i++ {
		envelopes[i] = 1
		remainingCount -= 1
	}

	// 分配剩余的钱
	for i := 0; i < int(num)-1; i++ {
		// 最大可分配金额
		maxCount := remainingCount / int64(num-int32(i)-1)

		// 随机分配金额
		amount := r.Int63n(maxCount + 1)

		// 分配金额大于0且不超过剩余金额
		envelopes[i] += amount
		remainingCount -= amount
	}

	// 最后一个红包拿到剩余的所有金额
	envelopes[num-1] = remainingCount

	return envelopes
}

// RandByWeight 随机权重算法
func RandByWeight(weightList []int32) int {
	if len(weightList) == 0 {
		return 0
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	total := int32(0)
	accumulate := make([]int32, len(weightList))
	for i, w := range weightList {
		total += w
		accumulate[i] = total
	}

	v := r.Int31n(total)
	for i, w := range accumulate {
		if v <= w {
			return i
		}
	}

	return 0
}
