package xlsx

import (
	"errors"
	"github.com/xuri/excelize/v2"
	"io"
	"reflect"
)

func ParseSheet(file *excelize.File, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("not ptr")
	}
	val = val.Elem()
	if val.Kind() != reflect.Slice {
		return errors.New("not slice")
	}
	itemType := val.Type().Elem()
	if itemType.Kind() != reflect.Struct {
		return errors.New("not struct")
	}
	newItemValue := func() reflect.Value {
		return reflect.New(itemType).Elem()
	}
	putItem := func(v reflect.Value) {
		val.Set(reflect.Append(val, v))
	}
	var sheetName string
	if s, ok := newItemValue().Interface().(SheetName); ok {
		sheetName = s.SheetName()
	} else {
		sheetName = itemType.Name()
	}

	if sheetIndex, err := file.GetSheetIndex(sheetName); err != nil {
		return err
	} else if sheetIndex < 0 {
		return nil
	}
	fieldIndex := make(map[string]int)
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		alias := field.Tag.Get("column")
		switch alias {
		case "":
			fieldIndex[field.Name] = i
		case "-":
			continue
		default:
			fieldIndex[alias] = i
		}
	}
	if len(fieldIndex) == 0 {
		return errors.New("empty column struct")
	}
	sheetIndex := make(map[string]int)
	for i := 1; ; i++ {
		name, err := file.GetCellValue(sheetName, GetAxis(i, 1))
		if err != nil {
			return err
		}
		if name == "" {
			break
		}
		if _, ok := fieldIndex[name]; ok {
			sheetIndex[name] = i
		}
	}
	if len(sheetIndex) == 0 {
		return errors.New("sheet column empty")
	}
	for i := 2; ; i++ {
		var (
			notEmpty int
			item     = newItemValue()
		)
		for column, index := range sheetIndex {
			s, err := file.GetCellValue(sheetName, GetAxis(index, i))
			if err != nil {
				return err
			}
			if s == "" {
				continue
			}
			notEmpty++
			if err = string2Value(s, item.Field(fieldIndex[column])); err != nil {
				return err
			}
		}
		if notEmpty > 0 {
			putItem(item)
		} else {
			break
		}
	}
	return nil
}

// ParseAll 可以传多个model，每个model对应一个sheet，sheet名称和结构体中的SheetName方法对应
func ParseAll(r io.Reader, models ...interface{}) error {
	if len(models) == 0 {
		return errors.New("empty models")
	}
	file, err := Open(r)
	if err != nil {
		return err
	}
	defer file.Close()
	for i := 0; i < len(models); i++ {
		if err := ParseSheet(file, models[i]); err != nil {
			return err
		}
	}
	return nil
}
