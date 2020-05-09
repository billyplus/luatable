package xlsx

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type baseReader struct {
	name      string        // name of sheet config
	data      [][]string    // data from excel
	filter    string        // 用来筛选字段的条件
	keyCount  int           // key的数量，当keyCount=0时，表示是数组，大于0表示是对象（或map）
	keyNext   int           // 下行的key所在列
	keyIndex  int           // 当前前的key所在列
	filterRow int           // 过滤词所在行
	keyRow    int           // key所在行
	typeRow   int           // 类型所在行
	row       int           // 当前行
	col       int           // 当前列
	indent    int           // 缩进
	rowCount  int           // 行数
	colCount  int           // 列数
	builder   *bytes.Buffer // strings.Builder for building result string
	// builder     *strings.Builder // strings.Builder for building result string
	doneChan    chan bool // chan for emit value
	filterflags []bool    // 用来标记该列是否需要导出
	// validCol    []int            // 用来标记需要导出的列
	cellTypes  []cellType // 每列的数据类型
	errors     []error    // 用来记录产生的错误
	state      stateFunc  // // the next lexing function to enter
	filterFunc FilterFunc // 过滤器
	wg         sync.WaitGroup
}

// stateFunc represents the state of the reader as a function that returns the next state.
type stateFunc func(*baseReader) stateFunc

// NewBaseReader 创建一个Reader，用来读取excel 文件
func NewBaseReader(name string, src [][]string, filter string, keyCount,
	filterRow, keyRow, typeRow, firstRow int) Reader {
	r := &baseReader{
		name:      name,
		data:      src,
		filter:    filter,
		keyCount:  keyCount,
		filterRow: filterRow,
		keyRow:    keyRow,
		typeRow:   typeRow,
	}
	r.data[3][0] = "comment"
	r.rowCount = len(src)
	for i := 0; i < len(src); i++ {
		if len(r.data[i]) < 2 || r.data[i][1] == "" {
			r.rowCount = i
			break
		}
	}
	fmt.Println("row count is ", r.rowCount)
	r.colCount = len(src[keyRow])
	for i := 1; i < r.colCount; i++ {
		if r.data[keyRow][i] == "" {
			r.colCount = i
		}
	}
	r.keyCount = keyCount
	// fmt.Println("key count is ", r.keyCount)
	r.filterflags = make([]bool, r.colCount)
	r.cellTypes = make([]cellType, r.colCount)
	r.builder = new(bytes.Buffer)
	r.errors = make([]error, 0, 5)
	r.doneChan = make(chan bool)
	r.row = firstRow
	r.col = 0
	r.keyNext = 0
	r.indent = 0

	// 使用默认的过滤器
	r.filterFunc = DefaultFilterFunc
	// 进行其它初始化工作
	r.init()

	return r
}

// SetFilterFunc 设置自定义的过滤器
func (r *baseReader) SetFilterFunc(filterFunc FilterFunc) {
	r.filterFunc = filterFunc
	for i := 0; i < r.colCount; i++ {
		// 设置导出标记，用来判断每列是否需要导出
		r.filterflags[i] = r.filterFunc(r.data[r.filterRow][i], r.filter)
	}
}

func (r *baseReader) refleshValidCol() {
	// r.validCol = make([]int, r.colCount)
	// validNum := 0
	for i := 0; i < r.colCount; i++ {
		// 设置导出标记，用来判断每列是否需要导出
		r.filterflags[i] = r.filterFunc(r.data[r.filterRow][i], r.filter)
		fmt.Println("filter i=", i, r.filterflags[i])
		// if r.filterFunc(r.data[r.filterRow][i], r.filter) {
		// 	r.validCol[validNum] = i
		// 	validNum++
		// }
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ReadAll 将excel 转为lua字符串
func (r *baseReader) ReadAll() ([]byte, error) {
	validcol := 0
	fmt.Println("read all")
	for i := 0; i < r.colCount; i++ {
		if r.filterflags[i] {
			validcol++
			break
		}
	}
	if validcol == 0 {
		return nil, ErrNoContent
	}

	//开始处理表格
	if err := r.run(); err != nil {
		return nil, err
	}

	return r.builder.Bytes(), nil
}

func (r *baseReader) run() (err error) {
	defer func() {
		if e := recover(); e != nil {
			if ex, ok := e.(error); ok {
				err = errors.Wrap(ex, "ReadAll")
			} else {
				err = errors.Errorf("%+v", ex)
			}
		}
	}()
	fmt.Println("start to run")
	for r.state = readBeginOfFile; r.state != nil; {
		r.state = r.state(r)
	}
	return
}

func (r *baseReader) emit(value string) {

	r.builder.WriteString(value)
}

func (r *baseReader) emitKey() {
	r.builder.WriteString(r.data[r.keyRow][r.col])
}

func (r *baseReader) emitValue() {
	switch r.cellTypes[r.col] {
	case cellString:
		r.emitString()
	case cellBool:
		r.emitBool()
	case cellFloat:
		r.emitFloat()
	case cellInt:
		r.emitInt()
	default:
		r.emitRawValue()
	}
}

func (r *baseReader) emitString() {
	if len(r.data[r.row]) <= r.col {
		fmt.Printf("data is empty sheet=%s row=%d col=%d\n", r.name, r.row, r.col)
		r.builder.WriteString("''")
		return
	}
	v := r.data[r.row][r.col]
	if strings.HasPrefix(v, "Lang.") {
		r.builder.WriteString(v)
	} else {
		r.builder.WriteByte('\'')
		r.builder.WriteString(v)
		r.builder.WriteByte('\'')
	}
}

func (r *baseReader) emitBool() {
	v := r.data[r.row][r.col]

	switch strings.ToLower(v) {
	case "0", "false", "":
		r.builder.WriteString("false")
	default:
		r.builder.WriteString("true")
	}
}

func (r *baseReader) emitInt() {
	v := r.data[r.row][r.col]
	if v == "" {
		r.builder.WriteString("0")
	} else {
		// 尝试int64
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// 尝试float64
			fval, err := strconv.ParseFloat(v, 64)
			if err != nil {
				// 不是数字
				r.errorf("第%d行%d列数值不是数字", r.row, r.col)
			}
			val = int64(fval)
		}
		if val > 1e11 {
			r.builder.WriteString(fmt.Sprintf("%e", float64(val)))
		} else {
			r.builder.WriteString(strings.ToLower(strconv.FormatInt(val, 10)))
		}
	}
}

func (r *baseReader) emitFloat() {
	v := r.data[r.row][r.col]
	if v == "" {
		r.builder.WriteString("0.0")
	} else {
		// 尝试float64
		_, err := strconv.ParseFloat(v, 64)
		if err != nil {
			// 不是数字
			r.errorf("第%d行%d列数值不是数字", r.row, r.col)
		}
		r.builder.WriteString(strings.ToLower(v))
		// if val > 1e11 {
		// 	r.builder.WriteString(fmt.Sprintf("%e", val))
		// } else {
		// }
		// r.builder.WriteString(v)
	}
}

func (r *baseReader) emitNumericKey() {
	r.builder.WriteString("[")
	r.builder.WriteString(r.data[r.row][r.col])
	r.builder.WriteString("]")
}

func (r *baseReader) emitRawValue() {
	r.builder.WriteString(r.data[r.row][r.col])
}

func (r *baseReader) emitComment() {
	r.builder.WriteString("--[[")
	r.builder.WriteString(r.data[r.row][r.col])
	r.builder.WriteString("]]")
}

func (r *baseReader) emitIndent() {
	if r.indent > 0 {
		r.builder.WriteString(strings.Repeat("\t", r.indent))
	}
}

func (r *baseReader) errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf("表%s第%d行(共%d行)第%d列数据[%s]有问题\n", r.name, r.row, r.rowCount, r.col, r.data[r.row][r.col]) + fmt.Sprintf(format, args...)
	panic(errors.New(msg))
}

// func (r *baseReader)writeString(){

// }

func readBeginOfFile(r *baseReader) stateFunc {
	// r.emit(r.name)
	// fmt.Println("begin of file")
	r.emit("{\n")
	r.indent++
	return readBeginOfLine
}

func readBeginOfLine(r *baseReader) stateFunc {
	// fmt.Println("begin of line row=", r.row)
	// 跳过首单元格为空的行
	// for {
	if r.row >= r.rowCount {
		return readEndOfFile
	}
	if len(r.data[r.row]) < 2 || r.data[r.row][1] == "" {
		return readEndOfFile
		// r.row++
		// continue
	}
	// break
	// }
	// fmt.Println("row is ", r.row)
	// keycount = 0 表示是数组
	if r.keyCount == 0 {
		r.emitIndent()
		r.emit("{\n")
		r.indent++
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
				r.emitIndent()
				switch r.cellTypes[r.col] {
				case cellInt:
					r.emitNumericKey()
				case cellString:
					r.emitRawValue()
				default:
					r.errorf("invalid key col=%d type=%d", r.col, r.cellTypes[r.col])
				}

				r.emit("={\n")
				// r.indent++
			}
			// r.keyIndex++
			j++
		}
	}
	r.col = oldColumn
	r.indent++

	return readNext
}

// func readKeys(r *baseReader) stateFunc {

// 	return readBeginOfLine
// }

func readEndOfLine(r *baseReader) stateFunc {
	// keycount = 0 表示是数组
	r.indent--
	if r.keyCount == 0 {
		r.emitIndent()
		r.emit("},\n")
	} else {
		for i := r.keyNext; i < r.keyCount; i++ {
			r.emitIndent()
			r.emit("},\n")
			// if i != r.keyCount-1 {
			// 	r.emit("\n")
			// }
		}
	}
	// 最后一列不需要逗号
	// if r.row < r.rowCount-1 {
	// 	r.emit(",")
	// }
	// r.emit(",\n")

	// 重置col的位置
	r.col = 0
	r.row++
	if r.row >= r.rowCount {
		return readEndOfFile
	}
	return readBeginOfLine
}

func readEndOfFile(r *baseReader) stateFunc {
	r.indent--
	r.emitIndent()
	r.emit("}")
	return nil
}

func readNext(r *baseReader) stateFunc {
	// fmt.Printf("col = %v, colcount=%v\n", r.col, r.colCount)
	if r.filterflags[r.col] {
		switch r.cellTypes[r.col] {
		case cellComment:
			// r.emitComment()
		case cellInvalid:
		// case cellString:
		// 	r.emitString()
		default:
			r.emitIndent()
			r.emitKey()
			r.emit(" = ")
			r.emitValue()
			// if r.col < r.colCount-1 {
			// 	r.emit(",")
			// }
			r.emit(",\n")
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
func (r *baseReader) init() {
	for i := 0; i < r.colCount; i++ {
		// 设置导出标记，用来判断每列是否需要导出
		r.filterflags[i] = r.filterFunc(r.data[r.filterRow][i], r.filter)
		// 设置每列的数据类型
		t := r.data[r.typeRow][i]
		if ctype, ok := stringToCellType(t); ok {
			r.cellTypes[i] = ctype
		} else {
			if r.filterflags[i] {
				panic(errors.Errorf("字段类型不正确 表:%s 列:%d", r.name, i))
			}
		}
	}
}
