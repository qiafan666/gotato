package utils

import (
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
)

// Export multiple sheets to Excel
func ExportToExcel(data map[string][]interface{}, srcPath string) error {
	// 创建Excel文件
	file := excelize.NewFile()

	// 遍历每个sheet的数据
	for sheetName, sheetData := range data {
		// 创建sheet
		file.NewSheet(sheetName)
		file.DeleteSheet("Sheet1")
		// 写入数据到表格
		for row, rowData := range sheetData {
			if rowSlice, ok := rowData.([]interface{}); ok {
				for col, value := range rowSlice {
					cell := fmt.Sprintf("%s%d", getColumnLetter(col), row+1) // 获取列字母和行号
					file.SetCellValue(sheetName, cell, value)
				}
			}
		}

		// 为列添加名称
		columnCount := len(sheetData[0].([]interface{}))
		firstRowData, ok := sheetData[0].([]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		for col := 0; col < columnCount; col++ {
			columnName := fmt.Sprintf("%s%d", getColumnLetter(col), 1)
			file.SetCellValue(sheetName, columnName, firstRowData[col]) // 使用第一行作为列名
			file.SetColWidth(sheetName, columnName, columnName, 15)     // 设置列宽度
		}
	}

	// 保存Excel文件
	err := file.SaveAs(srcPath)
	if err != nil {
		return err
	}
	return nil
}

// 获取对应索引的列字母
func getColumnLetter(index int) string {
	letter := ""
	for index >= 26 {
		letter = string('A'+(index%26)) + letter
		index = (index / 26) - 1
	}
	letter = string('A'+index) + letter
	return letter
}
