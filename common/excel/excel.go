package excel

import (
	"github.com/xuri/excelize/v2"
)

type Excel struct {
	*excelize.File
}

func New(opts ...excelize.Options) *Excel {
	return &Excel{excelize.NewFile(opts...)}
}

func (e *Excel) AddSheet(sheetName string, header []string, rows *[][]any) error {
	index, err := e.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// 表头样式：加粗、灰底、居中、大号
	headerStyle, err := e.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D9D9D9"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "666666", Style: 1},
			{Type: "top", Color: "666666", Style: 1},
			{Type: "right", Color: "666666", Style: 1},
			{Type: "bottom", Color: "666666", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// 数据样式：居中、标准字号
	dataStyle, err := e.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 11,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
	}

	// 写入表头
	for i, h := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		e.SetCellValue(sheetName, cell, h)
		e.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 写入数据（每个 data[i] 占一行，复制 header 列数）
	for i, row := range *rows {
		for j, val := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			e.SetCellValue(sheetName, cell, val)
			e.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
	}

	e.SetActiveSheet(index)

	return nil
}
