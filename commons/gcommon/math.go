package gcommon

import (
	"fmt"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/constraints"
)

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

// ParsePrecByInteger 将精度整形数字转换为浮点数字符串
// 例: 4->0.0001, 2->0.01
func ParsePrecByInteger(prec int) string {
	// 计算 10 的 -prec 次方
	return decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(0 - prec))).String()
}

// ParseFloatStrToPercentStr 将一个浮点数字符串转换为百分比字符串，并根据指定的源精度进行截断
// 例: vStr="0.0012", srcPrec=4 -> "0.1200"
func ParseFloatStrToPercentStr(vStr string, srcPrec int) string {
	// 将字符串转换为 decimal.Decimal 类型
	v, err := decimal.NewFromString(vStr)
	if err != nil {
		// 如果转换失败，返回空字符串
		return ""
	}

	// 将该值乘以 100 转换为百分比
	v = v.Mul(decimal.NewFromInt(100))

	// 如果 srcPrec 大于等于 2，则截断到 srcPrec - 2 位小数
	if srcPrec >= 2 {
		v = v.Truncate(int32(srcPrec - 2))
		// 格式化输出字符串
		format := fmt.Sprintf("%%0.%df", srcPrec-2)
		return fmt.Sprintf(format, v.InexactFloat64())
	}

	// 如果 srcPrec 小于 2，直接返回字符串形式的百分比值
	return v.String()
}

// ParseFloatStrToPercentWithRound 将一个浮点数字符串转换为百分比字符串，并根据指定的精度进行四舍五入
// 例: vStr="0.0012", round=2 -> "0.12"
func ParseFloatStrToPercentWithRound(vStr string, round int32) string {
	// 将字符串转换为 decimal.Decimal 类型
	v, err := decimal.NewFromString(vStr)
	if err != nil {
		// 如果转换失败，返回空字符串
		return ""
	}

	// 将该值乘以 100 转换为百分比，并进行四舍五入
	return v.Mul(decimal.NewFromInt(100)).Round(round).String()
}

// StrToDecimal 将一个字符串转换为 decimal.Decimal 类型
// 例: vStr="123.45" -> decimal.Decimal(123.45)
func StrToDecimal(vStr string) decimal.Decimal {
	// 将字符串转换为 decimal.Decimal 类型
	d, err := decimal.NewFromString(vStr)
	if err != nil {
		// 如果转换失败，返回 decimal.Zero
		return decimal.Zero
	}
	return d
}
