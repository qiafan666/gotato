package utils

import "testing"

func TestExport(t *testing.T) {

	data := map[string][]interface{}{
		"Sheet1": { // Data for Sheet1
			[]interface{}{"Name", "Age", "City"},
			[]interface{}{"John Doe", 25, "New York"},
			[]interface{}{"Jane Smith", 30, "London"},
		},
		"Sheet2": { // Data for Sheet2
			[]interface{}{"Product", "Price"},
			[]interface{}{"Apple", 1.99},
			[]interface{}{"Orange", 0.99},
		},
	}

	srcPath := "./xlsx/output.xlsx"
	//dstPath := "output.xlsx"
	err := ExportToExcel(data, srcPath)
	if err != nil {
		return
	}

	// 下载Excel文件
	//ctx.Header("Content-Disposition", "attachment; filename=output.xlsx")
	//ctx.Header("Content-Type", "application/octet-stream")
	//ctx.SendFile(srcPath, dstPath)
}
