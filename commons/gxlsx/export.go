package xlsx

import (
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
)

// Export multiple sheets to Excel
// data: map[sheetName][]interface{} sheetName为sheet名称，[]interface{}为每行的数据，第一行作为列名
// srcPath: 保存路径
func Export(data map[string][]interface{}, srcPath string) (err error) {
	// 创建Excel文件
	file := excelize.NewFile()
	err = file.DeleteSheet("Sheet1")
	if err != nil {
		return err
	}

	// 遍历每个sheet的数据
	for sheetName, sheetData := range data {
		// 创建sheet
		_, err = file.NewSheet(sheetName)
		if err != nil {
			return err
		}

		// 为列添加名称
		columnCount := len(sheetData[0].([]interface{}))
		firstRowData, ok := sheetData[0].([]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		for col := 0; col < columnCount; col++ {
			columnName := fmt.Sprintf("%s%d", getColumnLetter(col), 1)
			err = file.SetCellValue(sheetName, columnName, firstRowData[col])
			if err != nil {
				return err
			}
			// 使用第一行作为列名
			err = file.SetColWidth(sheetName, columnName, columnName, 15)
			if err != nil {
				return err
			}
		}

		// 写入数据到表格
		for row, rowData := range sheetData {
			if rowSlice, ok := rowData.([]interface{}); ok {
				for col, value := range rowSlice {
					cell := fmt.Sprintf("%s%d", getColumnLetter(col), row+1) // 获取列字母和行号
					err = file.SetCellValue(sheetName, cell, value)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// 保存Excel文件
	err = file.SaveAs(srcPath)
	if err != nil {
		return err
	}
	return nil
}

// 获取对应索引的列字母
func getColumnLetter(index int) string {
	letter := ""
	for index >= 26 {
		letter = fmt.Sprintf("%c%s", 'A'+(index%26), letter)
		index = (index / 26) - 1
	}
	letter = fmt.Sprintf("%c%s", 'A'+index, letter)
	return letter
}
