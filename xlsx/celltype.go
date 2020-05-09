package xlsx

import "strconv"

type cellType int

// var (
// 	CellString cellType = "string"
// 	CellInt cellType = "int"

// )

const (
	cellInvalid cellType = iota
	cellComment
	cellString
	cellInt
	cellFloat
	cellBool
	cellRaw
	cellTable
	end // end of defines
)

var cells = [...]string{
	cellInvalid: "invalid",
	cellComment: "comment",
	cellString:  "string",
	cellInt:     "int",
	cellFloat:   "float",
	cellBool:    "bool",
	cellRaw:     "raw",
	cellTable:   "table",
}

var cellsMap map[string]cellType

func init() {
	cellsMap = make(map[string]cellType, end)
	for i, v := range cells {
		cellsMap[v] = cellType(i)
	}
}

// ToString 输出对应的cellType string
func (t cellType) ToString() string {
	s := ""
	if 0 <= t && t < cellType(len(cells)) {
		s = cells[t]
	}
	if s == "" {
		s = "type(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

// stringToCellType 将string转换成cellType
func stringToCellType(typeStr string) (cellType, bool) {
	t, isType := cellsMap[typeStr]
	return t, isType
}

func isIdent(t cellType) bool {
	return t > cellComment
}
