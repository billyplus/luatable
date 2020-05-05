package luatable

import (
	"unicode"
	"unicode/utf8"
)

const bom = 0xFEFF // byte order mark, only permitted as very first character

type scanner struct {
	src []byte

	ch         rune // current character
	offset     int  // pos of current character
	rdOffset   int  // read offset, next character
	lineOffset int  // line offset
}

func (s *scanner) Init(src []byte) {
	s.src = src
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.next()
	if s.ch == bom {
		s.next()
	}
}

func (s *scanner) Scan() (pos Pos, tok tokType, val string) {
	s.skipWhiteSpace()
	ch := s.ch
	pos = s.pos(s.offset)
	if isLetter(ch) {
		val = s.scanIdent()
		tok = tokIdent
	} else if isDigit(ch) {
		tok, val = s.scanNumber()
	} else {
		s.next()
		switch ch {
		case '\'':
			tok, val = s.scanString('\'')
		case '"':
			tok, val = s.scanString('"')
		case '-':
			tok, val = s.scanComment()
		case '[':
			tok = tokLBracket
			val = tok.ToString()
		case ']':
			tok = tokRBracket
			val = tok.ToString()
		case '{':
			tok = tokLBrace
			val = tok.ToString()
		case '}':
			tok = tokRBrace
			val = tok.ToString()
		case '=':
			tok = tokAssign
			val = tok.ToString()
		case ':':
			tok = tokColon
			val = tok.ToString()
		case eof:
			tok = tokEOF
			val = tok.ToString()
		default:
			tok = tokError
		}
	}
	return
}

func (s *scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			// if pre is \n
			s.lineOffset = s.offset
		}
		ch, w := rune(s.src[s.offset]), 1
		switch {
		case ch == 0:
		case ch > utf8.RuneSelf:
			// not ascII
			ch, w = utf8.DecodeLastRune(s.src[s.offset:])
			switch {
			case ch == bom && s.offset > 0:
			case ch == utf8.RuneError && w == 1:
				// invalid utf8
			}
		}
		s.ch = ch
		s.rdOffset += w
	} else {
		s.offset = len(s.src)
		s.ch = eof
	}
}

func (s *scanner) skipWhiteSpace() {
	for s.ch == ' ' || s.ch == ',' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *scanner) pos(p int) Pos {
	if p > len(s.src) {
		panic("illigel file position")
	}
	return Pos(p)
}

func (s *scanner) scanIdent() string {
	start := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[start:s.offset])
}

func (s *scanner) scanNumber() (tok tokType, val string) {
	start := s.offset
	// integer part
	for isDigit(s.ch) {
		s.next()
	}

	if s.ch != '.' {
		// integer
		tok = tokInt
	} else {
		// float
		tok = tokFloat
		s.next() // move forward
		for isDigit(s.ch) {
			s.next()
		}
	}
	val = string(s.src[start:s.offset])
	return
}

func (s *scanner) scanString(quoteRune rune) (tok tokType, val string) {
	// first quote was comsumed
	start := s.offset - 1
ScanStringLoop:
	for {
		s.next() // move forward
		if s.ch == quoteRune {
			s.next()
			tok = tokString
			break ScanStringLoop
		}
		if s.ch == '\\' {
			// 转义字符
			s.next()
			switch s.ch {
			case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '"', '\'':
				// continue
			case 'x':
				// \xNN: hex value
				for i := 0; i < 2; i++ {
					s.next()
					if !isHex(s.ch) {
						tok = tokError
						break ScanStringLoop
					}
				}
			default:
				if isOct(s.ch) {
					// \ddd: 0ct value
					for i := 1; i < 3; i++ {
						s.next()
						if !isOct(s.ch) {
							tok = tokError
							break ScanStringLoop
						}
					}
				} else {
					// unknow escape sequence
					tok = tokError
					break ScanStringLoop
				}
			}
		}
	}
	val = string(s.src[start:s.offset])
	return
}

func (s *scanner) scanComment() (tok tokType, val string) {
	start := s.offset - 1
	if s.ch != '-' {
		tok = tokError
	} else {
		tok = tokComment
		s.next()
		if s.ch == '[' {
			s.next()
			if s.ch == '[' {
				// block comments
				for {
					s.next()
					switch s.ch {
					case ']':
						s.next()
						if s.ch == ']' {
							s.next()
							goto exit
						}
					case eof:
						tok = tokError
						goto exit
					}
				}
			}
		}
		for s.ch != '\n' {
			s.next()
		}
	}
exit:
	val = string(s.src[start:s.offset])
	if val[len(val)-1] == '\r' {
		// drop '\r'
		val = val[:len(val)-1]
	}
	return
}

// isLetter reports whether r is a alphabet character
func isLetter(r rune) bool {
	if r > unicode.MaxASCII {
		return false
	}
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r >= utf8.RuneSelf && unicode.IsLetter(r)
}

func isHex(r rune) bool {
	return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func isOct(r rune) bool {
	return r >= '0' && r <= '7'
}
