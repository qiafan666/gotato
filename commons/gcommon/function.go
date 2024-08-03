package gcommon

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/go-version"
	"github.com/kataras/iris/v12"
	"github.com/qiafan666/gotato/commons"
	"github.com/qiafan666/gotato/commons/log"
	"github.com/qiafan666/gotato/commons/utils"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

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

// StructToMapWithFilter 筛选出非nil的字段，转换成map,跳过指定字段，json标签为空的字段，json标签为数据库字段
func StructToMapWithFilter(inputStruct interface{}, JumpString ...string) map[string]interface{} {

	structValue := reflect.ValueOf(inputStruct)
	structType := structValue.Type()

	resultMap := make(map[string]interface{}, structType.Len())

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
