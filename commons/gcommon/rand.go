package gcommon

import (
	"math/rand"
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
