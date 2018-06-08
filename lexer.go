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

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= l.length {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
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

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

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
	switch r := l.next(); {
	case isSpace(r):
		return lexSpace
	case r == '{':
		l.emit(tokLBrace)
	case r == '}':
		l.emit(tokRBrace)
	case isAlphabet(r):
		return lexIdent
	case isDigit(r):
		return lexNumber
	case r == '=':
		l.emit(tokAssign)
	case r == '"':
		return lexDoubleQuote
	case r == '/':
		return lexComment
	case r == eof:
		l.emit(tokEOF)
		return nil
	default:
		l.errorf("无效的字符")
		return nil
	}
	return lexAny
}

// lexNumber scan a number. number can be a int or float
func lexNumber(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isDigit(r):
		case r == '.':
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
		switch r := l.next(); {
		case isDigit(r):
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
		switch r := l.next(); {
		// case r == '_':
		case isDigit(r):
		case isAlphabet(r):
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
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("未闭合的字符串。")
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
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("未闭合的字符串。")
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
		if !isSpace(l.next()) {
			l.backup()
			break
		}
	}
	//l.emit(TokS)
	l.ignore()
	return lexAny
}

func lexComment(l *lexer) stateFn {
	switch l.next() {
	case '*':
	Loop:
		for {
			switch l.next() {
			case '*':
				r := l.next()
				if r == '/' {
					l.emit(tokComment)
					break Loop
				}
				if r != eof {
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
