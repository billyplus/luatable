package luatable

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

var logger log.Logger

func init() {
	logger = log.NewLogfmtLogger(os.Stdout)
	caller := log.Caller(4)
	logger = log.With(logger, "caller", caller)
	// logger.Log("msg", "start")
}

// Unmarshal 将字符串解析成对象、数组、或变量
func Unmarshal(data []byte, v interface{}) error {
	var d decoder
	d.init(data)
	return d.unmarshal(v)
}

type decoder struct {
	// lexer   Lexer
	scanner scanner

	src   []byte
	token tokType
	pos   Pos
	val   string
	// peek    bool
	// thirdTok   tokToken
	savedError error
}

func (d *decoder) init(data []byte) {
	d.scanner.Init(data)
	d.src = data
	// d.lexer = newLexer("lexer", data)
	// d.nextTok = d.lexer.NextToken()
	// d.thirdTok = d.lexer.NextToken()
	d.next()
}

func (d *decoder) next() {
	d.pos, d.token, d.val = d.scanner.Scan()
	// fmt.Println("token=", d.token.ToString(), "val=[", d.val, "]")
	for d.token == tokComment {
		d.pos, d.token, d.val = d.scanner.Scan()
	}
	if d.token == tokError {
		d.error("无效字符:[" + d.val + "]")
	}
}

func (d *decoder) unmarshal(v interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if ex, ok := e.(error); ok {
				err = errors.Wrap(ex, "unmarshal")
			} else {
				err = errors.Errorf("%+v", ex)
			}
		}
	}()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("需要传入指针参数")
	}
	rv = d.indirect(rv)
	d.value(rv)
	return d.savedError
}

func (d *decoder) value(v reflect.Value) {
	tok := d.token
	switch tok {
	case tokLBrace:
		d.next()
		if d.token == tokLBracket || d.token == tokIdent {
			d.object(v)
		} else if d.token == tokLBrace || d.token.IsLiteral() {
			// fmt.Println(d.token.ToString(), d.val)
			// is array
			d.array(v)
		}

	// case tokIdent:
	// 	d.error("无效的起始字符")
	// 	// d.object(v)
	// case tokInt:
	// 	if d.nextTok == tokAssign {
	// 		d.object(v)
	// 	}
	// 	fallthrough
	default:
		d.error("无效的起始字符")
		// d.literal(v)
	}
}

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
func (d *decoder) indirect(v reflect.Value) reflect.Value {
	for {
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.IsNil() {
			// fmt.Println("v is nil")
			v.Set(reflect.New(v.Type().Elem()))
		}
		// fmt.Println("v is not nil")

		v = v.Elem()
	}
	// fmt.Println("v type is ", v.Type().String())
	return v
}

func (d *decoder) array(v reflect.Value) {
	// fmt.Println("is array", d.token.ToString(), d.val)
	switch v.Kind() {
	case reflect.Interface, reflect.Slice:
		v.Set(reflect.ValueOf(d.arrayInterface()))
		// if v.NumMethod() == 0 {
		// 	//没有初始化的interface{}
		// }
	}
	d.atEOF()

}

func (d *decoder) object(v reflect.Value) {
	// fmt.Println("object")
	switch v.Kind() {
	case reflect.Interface, reflect.Map:
		// fmt.Println("1")
		v.Set(reflect.ValueOf(d.objectInterface()))
		// if v.NumMethod() == 0 {
		// 	// fmt.Println("2")
		// 	//没有初始化的interface{}
		// 	// fmt.Println("object:v type is ", v.Type().String())
		// }
	default:
		// fmt.Println(v.Kind())
	}
	d.atEOF()
}

func (d *decoder) atEOF() {
	d.next()
	if d.token != tokEOF {
		d.error("结尾字符过多，检查前面的括号封闭")
	}
	// fmt.Println(d.token.Type().ToString())
}

func (d *decoder) expect(tok tokType) {
	if tok != d.token {
		d.error("expect token: " + tok.ToString())
	}
	d.next()
}

func (d *decoder) valueInterface() interface{} {
	// fmt.Println("valueInterface ", d.token.ToString(), d.val)
	switch d.token {
	case tokLBrace:
		d.next()
		// fmt.Println("valueInterface ", d.token.ToString(), d.val)
		if d.token == tokRBrace {
			return nil
		}
		if d.token == tokIdent || d.token == tokLBracket {
			return d.objectInterface()
		}
		return d.arrayInterface()
	default:
		// fmt.Println("valueInterface default", d.token.ToString(), d.val)
		return d.literalInterface()
	}
}

func (d *decoder) arrayInterface() []interface{} {
	// fmt.Println("arrayInterface", d.token.ToString(), d.val)
	v := make([]interface{}, 0)
	// d.next()
	for {
		v = append(v, d.valueInterface())
		d.next()
		if d.token == tokRBrace {
			// d.next()
			break
		}
	}
	return v
}

func (d *decoder) objectInterface() map[string]interface{} {
	m := make(map[string]interface{})
	for {
		// fmt.Println("objectInterface", d.token.ToString(), d.val)
		key := ""
		tok := d.token
		val := d.val
		if tok == tokLBracket {
			// '[', numeric key
			d.next()
			// fmt.Println("objectInterface key", d.token.ToString(), d.val)
			if d.token != tokInt && d.token != tokString {
				d.error("failed to parse key of table")
			}
			key = d.val
			d.next()
			d.expect(tokRBracket)
		} else if tok == tokIdent {
			key = val
			d.next()
		} else {
			d.error("invalid key for table")
		}
		// fmt.Println("objectInterface expect =", d.token.ToString(), d.val)

		d.expect(tokAssign)

		// fmt.Println("key=", key)
		m[key] = d.valueInterface()
		// fmt.Println("val=", d.val)
		d.next()
		// fmt.Println(d.token)
		if d.token == tokRBrace {
			// d.next()
			break
		}
	}
	return m
}

func (d *decoder) literalInterface() interface{} {
	tok := d.token
	val := d.val
	switch tok {
	case tokInt:
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			d.error(err.Error())
		}
		return num
	case tokFloat:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			d.error(err.Error())
		}
		return f
	case tokString:
		if val == "" {
			return ""
		}
		str, err := strconv.Unquote(val)
		if err != nil {
			d.error(err.Error() + "    val=[" + val + "]")
		}
		return str
	case tokIdent:
		return val
	case tokBool:
		if val == "false" {
			return false
		} else if val == "true" {
			return true
		}
		d.error("不支持的布尔值" + val)
	default:
		d.error(fmt.Sprintf("不支持的类型: %s", tok.ToString()))
	}
	return nil
}

func simpleStr(str string) string {
	if len(str) > 50 {
		return fmt.Sprintf("%s ... %s", str[:25], str[len(str)-25:])
	}
	return str
}

func (d *decoder) error(msg string) {
	// str := d.val
	pos := d.pos - 20
	if pos < 0 {
		pos = 0
	}
	// for i := 0; i < 5; i++ {
	// 	if d.token == tokEOF {
	// 		break
	// 	}
	// 	d.next()
	// 	str = str + d.val
	// }
	end := int(d.pos) + 50
	if len(d.src) < end {
		end = len(d.src)
	}
	panic(fmt.Errorf("错误: %s\n位于%d, \"%s\"", msg, pos, d.src[pos:end]))
}
