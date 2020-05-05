package luatable

import (
	"fmt"
	"strconv"
)

// Token 用来描述一个token实例
type Token interface {
	// Filter(cond interface{}) bool // 根据条件过滤token
	Position() string // 返回position的字符串表达式
	String() string   // 返回token的字符串表达式
	Value() string    // 获取实际的值
	Type() tokType    // 获取token type
}

//Type 定义所用到的token类型
type tokType int

// Tokens
const (
	tokEOF tokType = iota
	tokError

	tokLiteralBegin
	tokComment
	tokIdent
	tokInt
	tokFloat
	tokString
	tokBool
	tokLiteralEnd

	tokOperatorBegin
	tokAssign   // =
	tokLBrace   // {
	tokRBrace   // }
	tokLBracket // [
	tokRBracket // ]
	tokComma    // ,
	tokPeriod   // .
	tokColon    // :
	tokOperatorEnd
	// BOL // Begin of list
	// EOL // End of list
	tokEnd
)

// tokList
var tokList = [...]string{
	tokEOF:      "eof",
	tokError:    "error",
	tokComment:  "comment",
	tokIdent:    "ident",
	tokInt:      "int",
	tokFloat:    "float",
	tokString:   "string",
	tokBool:     "bool",
	tokAssign:   "=",
	tokLBrace:   "{",
	tokRBrace:   "}",
	tokLBracket: "[",
	tokRBracket: "]",
	tokComma:    ",",
	tokPeriod:   ".",
	tokColon:    ":",
}

// tokMap 是string和token的映射
var tokMap map[string]tokType

func init() {
	tokMap = make(map[string]tokType, tokEnd)
	for i, t := range tokList {
		tokMap[t] = tokType(i)
	}
	// tokMap["table"] = RawString
}

// ToString 输出对应的token string
func (tok tokType) ToString() string {
	s := ""
	if 0 <= tok && tok < tokType(len(tokList)) {
		s = tokList[tok]
	}
	if s == "" {
		s = "type(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// stringToToken 将string转换成token
func stringToToken(toktyp string) (tokType, bool) {
	t, isType := tokMap[toktyp]
	return t, isType
}

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
//
func (tok tokType) IsLiteral() bool {
	return tokLiteralBegin < tok && tok < tokLiteralEnd
}

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

const eof = -1

type luaToken struct {
	typ  tokType
	pos  Pos
	val  string
	line int
}

func (t *luaToken) String() string {
	switch t.typ {
	case tokEOF:
		return "EOF"
	case tokError:
		return "Error:" + t.val
	}
	return t.val
}

func (t *luaToken) Value() string {
	return t.val
}

func (t *luaToken) Position() string {
	return fmt.Sprintf("第%v行第%v个字符", t.line, t.pos)
}

func (t *luaToken) Type() tokType {
	return t.typ
}
