package gcommon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/jinzhu/copier"
	"github.com/qiafan666/gotato/commons/gcast"
	"gorm.io/gorm"
	"math"
	"reflect"
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
		constraint, err := version.NewConstraint(strings.Join(andVers, ","))
		if err != nil {
			return false, err
		}
		if constraint.Check(semVer) {
			return true, nil
		}
	}
	return false, nil
}

// Struct2Map 筛选出非nil的字段，转换成map,跳过指定字段，json标签为空的字段，json标签为数据库字段
// JumpString 跳过指定字段 不解析第二层struct
func Struct2Map(inputStruct interface{}, JumpString ...string) map[string]interface{} {

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

// IsDBNotFound 判断是否为“记录未找到”错误
func IsDBNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsDBError 判断是否为数据库非空错误（即真正的数据库异常）
func IsDBError(err error) bool {
	return err != nil && !IsDBNotFound(err)
}

// FetchPageData 泛型分页查询数据，由 scope 控制 Offset 和 Limit
func FetchPageData[T any](db *gorm.DB, table string, scope func(*gorm.DB) *gorm.DB) ([]T, error) {
	var result []T

	query := db.Table(table)
	if scope != nil {
		query = query.Scopes(scope)
	}

	err := query.Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FetchAllData 泛型分页查询所有数据，分页大小固定为 1000，由 scope 控制过滤逻辑
func FetchAllData[T any](db *gorm.DB, table string, scope func(*gorm.DB) *gorm.DB) ([]T, error) {
	const pageSize = 1000
	var allData []T
	offset := 0

	for {
		pageScope := func(db *gorm.DB) *gorm.DB {
			q := db.Offset(offset).Limit(pageSize)
			if scope != nil {
				q = scope(q)
			}
			return q
		}

		batch, err := FetchPageData[T](db, table, pageScope)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		allData = append(allData, batch...)
		if len(batch) < pageSize {
			break
		}
		offset += pageSize
	}

	return allData, nil
}
