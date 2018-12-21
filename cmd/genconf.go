package main

import (
	excel "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/billyplus/luatable/xlsx"
)

func init() {

}

func iterateWorksheet(sheets map[string]*WorkSheet, configs []ExportConfig) {
	for name, sheet := range sheets {
		for _, conf := range configs {

		}
	}
}

func openXlsx(file string) (map[string]*WorkSheet, error) {
	xlsfile, err := excel.OpenFile(file)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*WorkSheet)
	// sname := xlsfile.GetSheetName(1)
	for _, sh := range xlsfile.GetSheetMap() {
		data := xlsfile.GetRows(sh)
		if checkValidSheet(data) {
			sheet := &WorkSheet{
				Type:       data[0][1],
				ServerPath: data[0][3],
				ClientPath: data[1][3],
				Data:       data[3:],
			}
			result[sh] = sheet
		}
	}
	return result, nil
}

func checkValidSheet(data [][]string) bool {
	typ := data[0][1]
	if len(data) < 9 {
		return false
	}
	if typ == "base" || typ == "tiny" {
		return true
	}
	return false
}

type WorkSheet struct {
	Type       string
	Data       [][]string
	ServerPath string
	ClientPath string
	Name       string
	KeyCount   int
}

func readerFatory(sheet WorkSheet, filter string) xlsx.Reader {
	switch sheet.Type {
	case "base":
		return xlsx.NewBaseReader(sheet.Name, sheet.Data, filter, sheet.KeyCount, 1, 3, 2, 4)
	case "tiny":
		return xlsx.NewTinyReader(sheet.Name, sheet.Data, filter, 1, 2, 3, 4)
	}
}
