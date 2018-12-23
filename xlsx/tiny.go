package xlsx

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type tinyReader struct {
	name        string           // name of sheet config
	data        [][]string       // data from excel
	filter      string           // 用来筛选字段的条件
	filterCol   int              // 过滤词所在行
	keyCol      int              // key所在行
	typeCol     int              // 类型所在行
	valueCol    int              // 值所在的列
	filterflags []bool           // 用来标记该列是否需要导出
	rowCount    int              // 总行数
	cellTypes   []cellType       // 每行的类型
	builder     *strings.Builder // strings.Builder for building result string
	filterFunc  FilterFunc       // 过滤器
}

func NewTinyReader(name string, src [][]string, filter string, filterCol, keyCol, typeCol, valueCol int) Reader {
	r := &tinyReader{
		name:      name,
		data:      src[1:],
		filter:    filter,
		filterCol: filterCol,
		keyCol:    keyCol,
		typeCol:   typeCol,
		valueCol:  valueCol,
	}
	r.builder = new(strings.Builder)
	r.filterFunc = DefaultFilterFunc

	r.init()
	return r
}

// SetFilterFunc 设置自定义的过滤器
func (tiny *tinyReader) SetFilterFunc(filterFunc FilterFunc) {
	tiny.filterFunc = filterFunc
}

func (tiny *tinyReader) ReadAll() (string, error) {
	// tiny.builder.WriteString(tiny.name)
	tiny.builder.WriteString("{")
	for i, row := range tiny.data {
		if tiny.filterFunc(row[tiny.filterCol], tiny.filter) {
			typ, ok := stringToCellType(row[tiny.typeCol])
			if !ok {
				return "", errors.Errorf("表%s第%d行未知的数据类型:%s", tiny.name, i, row[tiny.typeCol])
			}
			// write key
			tiny.builder.WriteString(row[tiny.keyCol])
			tiny.builder.WriteRune('=')
			// write value
			switch typ {
			case cellString:
				tiny.builder.WriteString(strconv.Quote(row[tiny.valueCol]))
			case cellBool:
				v := row[tiny.valueCol]

				switch strings.ToLower(v) {
				case "0", "false":
					tiny.builder.WriteString("false")
				default:
					tiny.builder.WriteString("true")
				}
			case cellInt:
				v := row[tiny.valueCol]
				if v == "" {
					tiny.builder.WriteString("0")
				} else {
					tiny.builder.WriteString(v)
				}
			case cellFloat:
				v := row[tiny.valueCol]
				if v == "" {
					tiny.builder.WriteString("0.0")
				} else {
					tiny.builder.WriteString(v)
				}
			default:
				tiny.builder.WriteString(row[tiny.valueCol])
			}
			// write end
			tiny.builder.WriteRune(',')
		}
	}
	tiny.builder.WriteRune('}')

	return tiny.builder.String(), nil
}

func (tiny *tinyReader) init() {

}
