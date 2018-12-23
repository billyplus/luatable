package xlsx

import (
	excel "github.com/tealeg/xlsx"
	"strconv"
)

// SheetToSlice 将Sheet转成二维数组
func SheetToSlice(sheet *excel.Sheet) ([][]string, error) {
	s := [][]string{}
	for _, row := range sheet.Rows {
		if row == nil {
			continue
		}
		r := []string{}
		for _, cell := range row.Cells {
			str, err := cell.FormattedValue()
			if err != nil {
				// Recover from strconv.NumError if the value is an empty string,
				// and insert an empty string in the output.
				if numErr, ok := err.(*strconv.NumError); ok && numErr.Num == "" {
					str = ""
				} else {
					return s, err
				}
			}
			r = append(r, str)
		}
		s = append(s, r)
	}
	return s, nil
}
