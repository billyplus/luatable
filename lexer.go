package luatable

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Lexer ...
type Lexer interface {
	NextToken() Token
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name   string  // the name of the input; used only for error reports
	input  string  // the string being scanned
	length int     //length of the input string
	state  stateFn // the next lexing function to enter
	start  Pos     // start position of this item
	pos    Pos     // current position in the input
	width  Pos     // width of last rune read from input
	line   int
	cur    rune       // cur rune to deal with
	tokens chan Token // channel of scanned items
}

// var spaceCharacters = []rune{' ', '\t', '\n', '\r', ','}

// newLexer creates a new lexer for the input string.
//initializes itself to lex a string and launches the state machine as a goroutine, returning the lexer itself and a channel of items.
func newLexer(name, input string) Lexer {
	l := &lexer{
		name:   name,
		input:  input,
		length: len(input),
		tokens: make(chan Token, 5),
	}
	go l.run()
	return l
}

// next move to the next rune in the input.
func (l *lexer) next() {
	if int(l.pos) >= l.length {
		l.width = 0
		l.cur = eof
		return
	}
	var w int
	l.cur, w = utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if l.cur == '\n' {
		l.line++
	}
	return
}

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
	for l.state = lexStart(l); l.state != nil; {
		l.state = l.state(l)
	}
	close(l.tokens)
}

// emit passes an item back to the client.
func (l *lexer) emit(t tokType) {
	if t != tokComment {
		l.tokens <- &luaToken{
			typ:  t,
			pos:  l.start,
			val:  l.input[l.start:l.pos],
			line: l.line}
	}
	l.start = l.pos //move to current pos
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- &luaToken{
		typ:  tokError,
		pos:  l.start,
		val:  fmt.Sprintf(format, args...),
		line: l.line}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) NextToken() Token {
	tok := <-l.tokens
	return tok
}

// // peek returns but does not consume the next rune in the input.
// func (l *lexer) peek() rune {
// 	r := l.next()
// 	l.backup()
// 	return r
// }

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width

	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// lexStart start from beginning
func lexStart(l *lexer) stateFn {

	return lexAny
}

//lexAny deal with patten
func lexAny(l *lexer) stateFn {
	// }
	switch l.next(); {
	case isSpace(l.cur):
		return lexSpace
	case l.cur == '{':
		l.emit(tokLBrace)
	case l.cur == '}':
		l.emit(tokRBrace)
	case isAlphabet(l.cur):
		return lexIdent
	case isDigit(l.cur):
		return lexNumber
	case l.cur == '=':
		l.emit(tokAssign)
	case l.cur == '"':
		return lexDoubleQuote
	case l.cur == '/':
		return lexComment
	case l.cur == eof:
		l.emit(tokEOF)
		return nil
	default:
		l.errorf("无效的字符:%s", l.input[l.start:l.pos])
		return nil
	}
	return lexAny
}

// lexNumber scan a number. number can be a int or float
func lexNumber(l *lexer) stateFn {
Loop:
	for {
		switch l.next(); {
		case isDigit(l.cur):
		case l.cur == '.':
			return lexFloat
		default:
			l.backup()
			l.emit(tokInt)
			break Loop
		}
	}
	return lexAny
}

// lexFloat scan a float number, start from '.'
func lexFloat(l *lexer) stateFn {
Loop:
	for {
		switch l.next(); {
		case isDigit(l.cur):
		default:
			l.backup()
			l.emit(tokFloat)
			break Loop
		}
	}
	return lexAny
}

// lexIndet scan a identity
func lexIdent(l *lexer) stateFn {
Loop:
	for {
		l.next()
		if isDigit(l.cur) || isAlphabet(l.cur) {
			continue
		}
		switch l.cur {
		case '_':
		case '-':
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			switch {
			case word == "true", word == "false":
				l.emit(tokBool)
			default:
				l.emit(tokIdent)
			}
			break Loop
		}
	}
	return lexAny
}

// lexSingleQuote scans a single quoted string.
func lexSingleQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next(); l.cur {
		case '\\':
			if l.next(); l.cur != eof && l.cur != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("未闭合的字符串:%s", l.input[l.start:l.pos])
		case '\'':
			break Loop
		}
	}
	l.emit(tokString)

	return lexAny
}

// lexDoubleQuote scans a double quoted string.
func lexDoubleQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next(); l.cur {
		case '\\':
			if l.next(); l.cur != eof && l.cur != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("未闭合的字符串:%s", l.input[l.start:l.pos])
		case '"':
			break Loop
		}
	}
	l.emit(tokString)

	return lexAny
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *lexer) stateFn {
	for {
		l.next()
		if !isSpace(l.cur) {
			l.backup()
			break
		}
	}
	//l.emit(TokS)
	l.ignore()
	return lexAny
}

func lexComment(l *lexer) stateFn {
	switch l.next(); l.cur {
	case '*':
	Loop:
		for {
			switch l.next(); l.cur {
			case '*':
				l.next()
				if l.cur == '/' {
					l.emit(tokComment)
					break Loop
				}
				if l.cur != eof {
					continue
				}
				fallthrough
			case eof:
				l.errorf("注释未闭合")
			}
		}
	}
	return lexAny
}

// isSpace reports whether r is a space character.
// space include \t \n \r ,
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == ','
}

// isAlphabet reports whether r is a alphabet character
func isAlphabet(r rune) bool {
	if r > unicode.MaxASCII {
		return false
	}
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isDigit reports whether r is digit
func isDigit(r rune) bool {
	if r > unicode.MaxASCII {
		return false
	}
	return r >= '0' && r <= '9'
}
