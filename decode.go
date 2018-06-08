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
func Unmarshal(data string, v interface{}) error {
	var d decoder
	d.init(data)
	return d.unmarshal(v)
}

type decoder struct {
	lexer   Lexer
	token   Token
	nextTok Token
	peek    bool
	// thirdTok   tokToken
	savedError error
}

func (d *decoder) init(data string) {
	d.lexer = newLexer("lexer", data)
	d.nextTok = d.lexer.NextToken()
	// d.thirdTok = d.lexer.NextToken()
	d.next()
}

func (d *decoder) next() {
	d.token = d.nextTok
	d.nextTok = d.lexer.NextToken()
	if d.token.Type() == tokError {
		d.error("token error")
	}
	// logger.Log("tok", d.token.Value())
	// for {
	// 	tok := d.lexer.NextToken()
	// 	if d.token.Type() == tokEOF || d.token.Type() == tokError || d.token.Type() != tokComment {
	// 		break
	// 	}
	// 	d.token = d.nextTok
	// 	d.nextTok = d.thirdTok
	// 	d.thirdTok = tok
	// }
	// d.next()
}

func (d *decoder) unmarshal(v interface{}) (err error) {
	defer func() {
		e := recover()
		if ex, ok := e.(error); ok {
			// fmt.Println(err.Error())
			err = errors.WithStack(ex)
			// err = ex
		}
		// fmt.Println(e)
	}()

	rv := reflect.ValueOf(v)
	rv = d.indirect(rv)
	d.value(rv)
	return d.savedError
}

func (d *decoder) value(v reflect.Value) {

	switch d.token.Type() {
	case tokLBrace:
		d.next()
		if d.token.Type() == tokIdent {
			d.object(v)
			break
		} else if d.token.Type() == tokInt || d.nextTok.Type() == tokAssign {
			d.object(v)
			break
		}

		d.array(v)
	case tokIdent:
		d.error("无效的起始字符")
		// d.object(v)
	case tokInt:
		if d.nextTok.Type() == tokAssign {
			d.object(v)
		}
		fallthrough
	default:
		d.literal(v)
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
	switch v.Kind() {
	case reflect.Interface:
		if v.NumMethod() == 0 {
			//没有初始化的interface{}
			v.Set(reflect.ValueOf(d.arrayInterface()))
		}
	}
	d.atEOF()

}

func (d *decoder) object(v reflect.Value) {
	// fmt.Println("object")
	switch v.Kind() {
	case reflect.Interface:
		// fmt.Println("1")
		if v.NumMethod() == 0 {
			// fmt.Println("2")
			//没有初始化的interface{}
			v.Set(reflect.ValueOf(d.objectInterface()))
			// fmt.Println("object:v type is ", v.Type().String())
		}
	default:
		fmt.Println(v.Kind())
	}
	d.atEOF()
}

func (d *decoder) atEOF() {
	d.next()
	if d.token.Type() != tokEOF {
		d.error("结尾字符过多，检查前面的括号封闭")
	}
	// fmt.Println(d.token.Type().ToString())
}

func (d *decoder) literal(v reflect.Value) {

}
func (d *decoder) valueInterface() interface{} {
	// fmt.Println("valueInterface")

	switch d.token.Type() {
	case tokLBrace:
		d.next()
		if d.token.Type() == tokRBrace {
			return nil
		}
		if d.token.Type() == tokIdent {
			return d.objectInterface()
		} else if d.token.Type() == tokInt || d.nextTok.Type() == tokAssign {
			return d.objectInterface()
		}
		return d.arrayInterface()
	case tokIdent:
		return d.objectInterface()
	case tokInt:
		if d.nextTok.Type() == tokAssign {
			return d.objectInterface()
		}
		fallthrough
	default:
		return d.literalInterface()
	}
}

func (d *decoder) arrayInterface() []interface{} {
	v := make([]interface{}, 0)
	// d.next()
	for {
		v = append(v, d.valueInterface())
		d.next()
		if d.token.Type() == tokRBrace {
			// d.next()
			break
		}
	}
	return v
}

func (d *decoder) objectInterface() map[string]interface{} {
	// fmt.Println("objectInterface")
	m := make(map[string]interface{})
	// d.next()
	for {
		key := d.token.Value()
		d.next()
		if d.token.Type() != tokAssign {
			d.error("缺少=号")
		}
		d.next()
		m[key] = d.valueInterface()
		d.next()
		// fmt.Println(d.token)
		if d.token.Type() == tokRBrace {
			// d.next()
			break
		} else if d.token.Type() == tokLBrace {
			d.error("缺少}号")
		}
	}
	return m
}

func (d *decoder) literalInterface() interface{} {
	v := d.token.Value()
	switch d.token.Type() {
	case tokInt:
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			d.error(err.Error())
		}
		return num
	case tokFloat:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			d.error(err.Error())
		}
		return f
	case tokString:
		str, err := strconv.Unquote(v)
		if err != nil {
			d.error(err.Error())
		}
		return str
	default:
		d.error("不支持的类型")
	}
	return nil
}

func (d *decoder) stringInterface() string {
	return d.token.Value()
}

func (d *decoder) error(msg string) {
	str := d.token.Value()
	for i := 0; i < 5; i++ {
		if d.token.Type() == tokEOF {
			break
		}
		d.next()
		str = str + d.token.Value()
	}
	panic(fmt.Errorf("位于%v前面:%v", str, msg))
}
