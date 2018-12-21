package xlsx

import "strconv"

type SheetType int

// var (
// 	CellString SheetType = "string"
// 	CellInt SheetType = "int"

// )

const (
	SheetInvalid SheetType = iota
	SheetBase
	SheetTiny
	SheetEnd // end of defines
)

var sheets = [...]string{
	SheetInvalid: "invalid",
	SheetBase:    "base",
	SheetTiny:    "tiny",
}

var sheetMap map[string]SheetType

func init() {
	sheetMap = make(map[string]SheetType, end)
	for i, v := range cells {
		sheetMap[v] = SheetType(i)
	}
}

// ToString 输出对应的SheetType string
func (t SheetType) ToString() string {
	s := ""
	if 0 <= t && t < SheetType(len(sheets)) {
		s = sheets[t]
	}
	if s == "" {
		s = "type(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

// StringToSheetType 将string转换成SheetType
func StringToSheetType(typeStr string) (SheetType, bool) {
	t, isType := sheetMap[typeStr]
	return t, isType
}
