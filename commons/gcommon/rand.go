package gcommon

import (
	"math/rand"
	"time"
)

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
	if num <= 0 {
		return nil
	}
	envelopes := make([]int64, num)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	remainingCount := totalCount
	for i := 0; i < int(num)-1; i++ {
		maxCount := (remainingCount / (int64(num) - int64(i))) * 2
		amount := r.Int63n(maxCount + 1)
		envelopes[i] = amount
		remainingCount -= amount
	}
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
