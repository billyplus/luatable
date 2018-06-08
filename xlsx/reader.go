package xlsx

import (
	"fmt"
	"strings"
	"sync"

	"github.com/billyplus/luasheet/reader"
)

type xlsxReader struct {
	name        string           // name of sheet config
	data        [][]string       // data from excel
	filter      string           // 用来筛选字段的条件
	keyCount    int              // key的数量，当keyCount=0时，表示是数组，大于0表示是对象（或map）
	keyNext     int              // 下行的key所在列
	keyIndex    int              // 当前前的key所在列
	filterRow   int              // 过滤词所在行
	keyRow      int              // key所在行
	typeRow     int              // 类型所在行
	row         int              // 当前行
	col         int              // 当前列
	rowCount    int              // 行数
	colCount    int              // 列数
	builder     *strings.Builder // strings.Builder for building result string
	doneChan    chan bool        // chan for emit value
	filterflags []bool           // 用来标记该列是否需要导出
	cellTypes   []cellType       // 每列的数据类型
	errors      []error          // 用来记录产生的错误
	state       stateFunc        // // the next lexing function to enter
	wg          sync.WaitGroup
}

// stateFunc represents the state of the reader as a function that returns the next state.
type stateFunc func(*xlsxReader) stateFunc

// New 创建一个Reader，用来读取excel 文件
func New(name string, src [][]string, filter string, keyCount int, filterRow int, keyRow int, typeRow int, firstRow int) reader.Reader {
	r := &xlsxReader{
		name:      name,
		data:      src,
		filter:    filter,
		keyCount:  keyCount,
		filterRow: filterRow,
		keyRow:    keyRow,
		typeRow:   typeRow,
	}
	r.rowCount = len(src)
	fmt.Println("row count is ", r.rowCount)
	r.colCount = len(src[0])
	r.keyCount = keyCount
	fmt.Println("key count is ", r.keyCount)
	r.filterflags = make([]bool, r.colCount)
	r.cellTypes = make([]cellType, r.colCount)
	r.builder = new(strings.Builder)
	r.errors = make([]error, 0, 5)
	r.doneChan = make(chan bool)
	r.row = firstRow
	r.col = 0
	r.keyNext = 0
	// 进行其它初始化工作
	r.init()

	//开始处理表格
	go r.run()
	return r
}

func (r *xlsxReader) ReadAll() (string, error) {
	<-r.doneChan
	// b := []byte(r.builder.String())
	var err error
	if len(r.errors) == 0 {
		err = nil
	} else {
		var errStr strings.Builder
		for _, e := range r.errors {
			errStr.WriteString("____")
			errStr.WriteString(e.Error())
			errStr.WriteByte('\n')
		}
		err = fmt.Errorf(errStr.String())
	}

	return r.builder.String(), err
}

func (r *xlsxReader) run() {
	for r.state = readBeginOfFile; r.state != nil; {
		r.state = r.state(r)
	}
	r.done()
}

func (r *xlsxReader) done() {
	r.doneChan <- true
}

func (r *xlsxReader) emit(value string) {

	r.builder.WriteString(value)
}

func (r *xlsxReader) emitKey() {
	r.builder.WriteString(r.data[r.keyRow][r.col])
}

func (r *xlsxReader) emitValue() {
	switch r.cellTypes[r.col] {
	case cellString:
		r.emitString()
	default:
		r.emitRawValue()
	}
}

func (r *xlsxReader) emitString() {
	r.builder.WriteByte('"')
	r.builder.WriteString(r.data[r.row][r.col])
	r.builder.WriteByte('"')
}

func (r *xlsxReader) emitRawValue() {
	r.builder.WriteString(r.data[r.row][r.col])
}

func (r *xlsxReader) emitComment() {
	r.builder.WriteString("/*")
	r.builder.WriteString(r.data[r.row][r.col])
	r.builder.WriteString("*/")
}

func (r *xlsxReader) errorf(format string, args ...interface{}) {
	r.errors = append(r.errors, fmt.Errorf(format, args...))
}

// func (r *xlsxReader)writeString(){

// }

func readBeginOfFile(r *xlsxReader) stateFunc {
	r.emit(r.name)
	r.emit("={")
	return readBeginOfLine
}

func readBeginOfLine(r *xlsxReader) stateFunc {

	// fmt.Println("row is ", r.row)
	// keycount = 0 表示是数组
	if r.keyCount == 0 {
		r.emit("{")
		return readNext
	}

	oldColumn := r.col
	r.keyIndex = r.keyNext
	j := 0
	// 查找下一个数据key开始的位置
	if r.row >= r.rowCount-1 {
		r.keyNext = 0
	} else {
		for i := 0; i < r.colCount; i++ {
			if isIdent(r.cellTypes[i]) {
				cur := r.data[r.row][i]
				next := r.data[r.row+1][i]
				if cur != next {
					r.keyNext = j
					break
				}
				j++
			}
		}
	}
	// fmt.Printf("index=%v,next=%v\n", r.keyIndex, r.keyNext)
	j = 0
	for r.col = 0; r.col <= r.colCount; r.col++ {

		if j >= r.keyCount {
			break
		}
		if isIdent(r.cellTypes[r.col]) {
			if j >= r.keyIndex {
				r.emitRawValue()
				r.emit("={")
			}
			// r.keyIndex++
			j++
		}
	}
	r.col = oldColumn

	return readNext
}

// func readKeys(r *xlsxReader) stateFunc {

// 	return readBeginOfLine
// }

func readEndOfLine(r *xlsxReader) stateFunc {
	// keycount = 0 表示是数组
	if r.keyCount == 0 {
		r.emit("}")
	} else {
		for i := r.keyNext; i < r.keyCount; i++ {
			r.emit("}")
		}
	}
	// 最后一列不需要逗号
	if r.row < r.rowCount-1 {
		r.emit(",")
	}

	// 重置col的位置
	r.col = 0
	r.row++
	if r.row >= r.rowCount {
		return readEndOfFile
	}
	return readBeginOfLine
}

func readEndOfFile(r *xlsxReader) stateFunc {
	r.emit("}\n")
	return nil
}

func readNext(r *xlsxReader) stateFunc {
	// fmt.Printf("col = %v, colcount=%v\n", r.col, r.colCount)
	switch r.cellTypes[r.col] {
	case cellComment:
		// r.emitComment()
	case cellInvalid:
	// case cellString:
	// 	r.emitString()
	default:
		r.emitKey()
		r.emit("=")
		r.emitValue()
		if r.col < r.colCount-1 {
			r.emit(",")
		}
	}

	r.col++
	// fmt.Printf("col = %v, colcount=%v\n", r.col, r.colCount)
	if r.col >= r.colCount {
		return readEndOfLine(r)
	}
	return readNext
}

// 在初始化时
// 设置标记每一列是否需要导出，这样就不需要每行都决断一次
// 设置每列的数据类型
func (r *xlsxReader) init() {
	for i := 0; i < r.colCount; i++ {
		// 设置导出标记，用来判断每列是否需要导出
		if strings.Contains(r.data[r.filterRow][i], r.filter) {
			r.filterflags[i] = true
		}
		// 设置每列的数据类型
		t := r.data[r.typeRow][i]
		if ctype, ok := stringToCellType(t); ok {
			r.cellTypes[i] = ctype
		} else {
			r.cellTypes[i] = cellInvalid
		}
	}
}
