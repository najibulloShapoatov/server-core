package settings

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type token int

const (
	illegal token = iota
	eof
	eol
	number
	float
	str
	duration
	comment
	ident
	brace_start
	brace_end
	variable
	equal
)

const bom = 0xFEFF

type parser struct {
	r          *bytes.Reader // byte reader
	offset     int           // read offset
	ch         rune          // current character
	line       int           // line number
	lineOffset int           // position where the new line begins
	err        error         // last read error
	filename   string
}

func (p *parser) error(msg string) {
	p.err = fmt.Errorf("%s (%s:%d:%d)",
		msg, p.filename, p.line+1, int(math.Max(0, float64(p.offset-p.lineOffset-1))))
}

func (p *parser) scan() (tok token, lit string) {
	p.skipWhitespace()
	switch ch := p.ch; {
	case isLetter(ch):
		lit = p.scanIdentifier()
		tok = ident
	case isDigit(ch) || (ch == '-' && isDigit(p.peek())):
		tok, lit = p.scanNumber()
	default:
		p.next()
		switch ch {
		case -1:
			tok = eof
		case '=':
			tok = equal
			lit = "="
		case '\n':
			tok = eol
			lit = "\n"
		case '#':
			tok = comment
			lit = p.scanComment()
		case '"':
			tok = str
			lit = p.scanString(false)
		case '}':
			tok = brace_end
		case '{':
			tok = brace_start
		case '$':
			tok = variable
			p.next()
			lit = fmt.Sprintf("${%s}", p.scanString(true))
		default:
			if ch != bom {
				p.error(fmt.Sprintf("illegal character %#U", ch))
			}
			tok = illegal
			lit = string(ch)
		}
	}
	return
}

func (p *parser) scanEscape() ([]byte, bool) {
	offset := p.offset
	var n int
	var base, max uint32
	switch p.ch {
	case 'a', 'b', 'n', 'r', 't', 'v', '\\', '"':
		p.next()
		switch p.ch {
		case 'a':
			p.next()
			return []byte("\a"), true
		case 'b':
			p.next()
			return []byte("\b"), true
		case 'n':
			p.next()
			return []byte("\n"), true
		case 'r':
			p.next()
			return []byte("\r"), true
		case 't':
			p.next()
			return []byte("\t"), true
		case 'v':
			p.next()
			return []byte("\v"), true
		case '\\':
			p.next()
			return []byte{'\\'}, true
		case '"':
			p.next()
			return []byte{'"'}, true
		}
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		p.next()
		n, base, max = 2, 16, 255
	case 'u':
		p.next()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		p.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		p.error("unknown escape sequence")
		if p.ch < 0 {
			p.error("escape sequence not terminated")
		}
		return nil, false
	}
	var x uint32
	for n > 0 {
		d := uint32(digitVal(p.ch))
		if d > base {
			p.error(fmt.Sprintf("illegal character %#U in escape sequence", p.ch))
			if p.ch < 0 {
				p.error("escape sequence not terminated")
			}
			return nil, false
		}
		x = x*base + d
		p.next()
		n--
	}
	if x > max || 0xd800 <= x && x < 0xe000 {
		p.error("escape sequence is invalid Unicode code point")
		return nil, false
	}
	var b = make([]byte, p.offset-offset)
	_, _ = p.r.ReadAt(b, int64(offset))
	return b, true
}

func (p *parser) scanString(isVariable bool) string {
	// " already consumed
	var buf = &bytes.Buffer{}
	for {
		if (isVariable && (p.ch == '{' || p.ch == '}')) || p.ch == '"' {
			p.next()
			break
		}
		if p.ch == '\\' {
			if peek := p.peek(); peek == '\r' || peek == '\n' {
				if peek == '\r' {
					p.next()
				}
				p.next()
				buf.WriteRune(p.ch)
				p.next()
				continue
			} else {
				if b, ok := p.scanEscape(); ok {
					buf.Write(b)
					continue
				} else {
					break
				}
			}
		}
		if p.ch == '\n' || p.ch < 0 {
			p.error("string literal not terminated")
			break
		}

		buf.WriteRune(p.ch)
		p.next()
	}
	return buf.String()
}

func (p *parser) scanNumber() (token, string) {
	offset := p.offset
	tok := illegal

	base := 10
	prefix := rune(0)
	times := 1

	if p.ch == '-' {
		times = -1
		p.next()
		offset = p.offset
	}

	// integer part
	if p.ch != '.' {
		tok = number
		if p.ch == '0' {
			p.next()
			switch lower(p.ch) {
			case 'x':
				p.next()
				base, prefix = 16, 'x'
				offset = p.offset
			case 'o':
				p.next()
				base, prefix = 8, 'o'
				offset = p.offset
			case 'b':
				p.next()
				base, prefix = 2, 'b'
				offset = p.offset
			default:
				base, prefix = 8, 'o'
				offset = p.offset
			}
		}
		p.digits(base)
	}

	// fractional part
	if p.ch == '.' {
		tok = float
		if prefix == 'o' || prefix == 'b' {
			p.error("invalid radix point")
		}
		p.next()
		p.digits(base)
	}

	// exponent
	if e := lower(p.ch); e == 'e' || e == 'p' {
		switch {
		case e == 'e' && prefix != 0 && prefix != '0':
			p.error(fmt.Sprintf("%q exponent requires decimal mantissa", p.ch))
		case e == 'p' && prefix != 'x':
			p.error(fmt.Sprintf("%q exponent requires hexadecimal mantissa", p.ch))
		}
		p.next()
		tok = float
		if p.ch == '+' || p.ch == '-' {
			p.next()
		}
		p.digits(10)
	} else if prefix == 'x' && tok == float {
		p.error("hexadecimal mantissa requires a 'p' exponent")
	}

	var (
		val interface{}
		err error
		b   = make([]byte, p.offset-offset)
	)

	_, _ = p.r.ReadAt(b, int64(offset)-1)
	lit := string(b)

	if tok == float {
		f, e := strconv.ParseFloat(lit, base)
		if e != nil {
			err = fmt.Errorf("error parsing float")
		}
		val = float64(times) * f
	} else {
		f, e := strconv.ParseInt(lit, base, 64)
		if e != nil {
			err = fmt.Errorf("error parsing int")
		}
		val = int64(times) * f
	}
	if err != nil {
		p.error(err.Error())
	}
	lit = fmt.Sprintf("%v", val)
	return tok, lit
}

func (p *parser) digits(base int) {
	if base <= 10 {
		for '0' <= p.ch && p.ch <= '9' {
			p.next()
		}
	} else {
		for '0' <= p.ch && p.ch <= '9' || 'a' <= lower(p.ch) && lower(p.ch) <= 'p' {
			p.next()
		}
	}
}

func (p *parser) scanComment() string {
	var buf = &bytes.Buffer{}
	p.next()
	for p.ch != '\n' && p.ch >= 0 {
		if p.ch != '\r' {
			buf.WriteRune(p.ch)
		}
		p.next()
	}
	return buf.String()
}

func (p *parser) scanIdentifier() string {
	var buf = &bytes.Buffer{}
	for isLetter(p.ch) || isDigit(p.ch) || isSymbol(p.ch) || p.ch == '\\' {
		if p.ch == '\\' {
			runes, ok := p.scanEscape()
			if !ok {
				return ""
			}
			buf.Write(runes)
		} else {
			buf.WriteRune(p.ch)
			p.next()
		}
	}
	return buf.String()
}

func (p *parser) skipWhitespace() {
	for p.ch == ' ' || p.ch == '\t' || p.ch == '\n' || p.ch == '\r' {
		p.next()
	}
}

func (p *parser) next() {
	ch, width, err := p.r.ReadRune()
	if err == io.EOF {
		p.ch = -1
		return
	}
	p.ch = ch
	p.offset += width

	if ch == '\n' {
		p.line++
		p.lineOffset = p.offset
	}
}

func (p *parser) peek() rune {
	ch, _, _ := p.r.ReadRune()
	_ = p.r.UnreadRune()
	return ch
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func isSymbol(ch rune) bool {
	return ch == '.' || ch == '_' || ch == '-'
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16
}

func lower(ch rune) rune {
	return ('a' - 'A') | ch
}
