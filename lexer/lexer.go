package lexer

import (
	"llvm-lang/token"
	"llvm-lang/utils"
)

type Lexer struct {
	source       string
	position     int
	readPosition int
	char         byte
}

const (
	eqSym = '='

	leftParen          = '('
	rightParen         = ')'
	leftSquareBracket  = '['
	rightSquareBracket = ']'
	leftCurlyBracket   = '{'
	rightCurlyBracket  = '}'

	semi  = ';'
	comma = ','
	colon = ':'
	dot   = '.'
	quote = '"'

	plus   = '+'
	star   = '*'
	slash  = '/'
	minus  = '-'
	modulo = '%'

	greaterThan = '>'
	lessThan    = '<'
	bang        = '!'
	ampersand   = '&'
	pipe        = '|'
)

var keywords = map[string]token.TokenType{
	"def":    token.Def,
	"extern": token.Extern,
}

func New(source string) *Lexer {
	lexer := &Lexer{source: source} // Start our lexer at line 1
	lexer.readChar()                // set up lexer
	return lexer
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.source) {
		l.char = 0
	} else {
		l.char = l.source[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) readIdentifer() string {
	position := l.position

	for utils.IsAlpha(l.char) {
		l.readChar() // Advances the position pointer
	}
	return l.source[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for utils.IsNumeric(l.char) || l.char == dot {
		l.readChar() // This just advances the position pointer
	}
	return l.source[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1 // advance past ""

	for {
		l.readChar()
		if l.char == quote || l.char == 0 {
			break
		}
	}
	return l.source[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.char == ' ' || l.char == '\t' || l.char == '\n' || l.char == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.source) {
		return 0
	} else {
		return l.source[l.readPosition]
	}
}

func LookupIdent(ident string) token.TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return token.Identifier
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()
	switch l.char {
	// grouping
	case leftParen:
		tok = token.MakeToken(token.LeftParen, l.char)
	case rightParen:
		tok = token.MakeToken(token.RightParen, l.char)
	case leftCurlyBracket:
		tok = token.MakeToken(token.LeftCurlyBracket, l.char)
	case rightCurlyBracket:
		tok = token.MakeToken(token.RightCurlyBracket, l.char)
	case leftSquareBracket:
		tok = token.MakeToken(token.LeftSquareBracket, l.char)
	case rightSquareBracket:
		tok = token.MakeToken(token.RightSquareBracket, l.char)

	// Punctuation
	case semi:
		tok = token.MakeToken(token.Semicolon, l.char)
	case comma:
		tok = token.MakeToken(token.Comma, l.char)
	case colon:
		tok = token.MakeToken(token.Colon, l.char)
	case dot:
		tok = token.MakeToken(token.Dot, l.char)
	case quote:
		tok.Type = token.String
		tok.Literal = l.readString()
	// Symbols
	case eqSym:
		if l.peekChar() == eqSym {
			char := l.char
			l.readChar() // advance past first equals
			literal := string(char) + string(l.char)
			tok = token.Token{Type: token.EqualTo, Literal: literal}
		} else {
			tok = token.MakeToken(token.Assign, l.char)
		}
	case plus:
		tok = token.MakeToken(token.Plus, l.char)
	case minus:
		tok = token.MakeToken(token.Minus, l.char)
	case star:
		tok = token.MakeToken(token.Star, l.char)
	case slash:
		tok = token.MakeToken(token.Slash, l.char)
	case modulo:
		tok = token.MakeToken(token.Modulo, l.char)
	case greaterThan:
		if l.peekChar() == eqSym {
			char := l.char
			l.readChar() // advance past first equals
			literal := string(char) + string(l.char)
			tok = token.Token{Type: token.GreaterThanEqualTo, Literal: literal}
		} else {
			tok = token.MakeToken(token.GreaterThan, l.char)
		}
	case lessThan:
		if l.peekChar() == eqSym {
			char := l.char
			l.readChar() // advance past first equals
			literal := string(char) + string(l.char)
			tok = token.Token{Type: token.LessThanEqualTo, Literal: literal}
		} else {
			tok = token.MakeToken(token.LessThan, l.char)
		}
	case bang:
		if l.peekChar() == eqSym {
			char := l.char
			l.readChar() // advance past first equals
			literal := string(char) + string(l.char)
			tok = token.Token{Type: token.NotEqualTo, Literal: literal}
		} else {
			tok = token.MakeToken(token.Bang, l.char)
		}
	case ampersand:
		if l.peekChar() == ampersand {
			char := l.char
			l.readChar()
			literal := string(char) + string(l.char)
			tok = token.Token{Type: token.And, Literal: literal}
		} else {
			// Single & is an illegal char
			tok = token.MakeToken(token.Illegal, l.char)
		}
	case pipe:
		if l.peekChar() == pipe {
			char := l.char
			l.readChar()
			literal := string(char) + string(l.char)
			tok = token.Token{Type: token.Or, Literal: literal}
		} else {
			// Single & is an illegal char
			tok = token.MakeToken(token.Illegal, l.char)
		}
	case 0:
		tok.Literal = ""
		tok.Type = "EOF"

	default:
		if utils.IsAlpha(l.char) {
			tok.Literal = l.readIdentifer()
			tok.Type = LookupIdent(tok.Literal)
			return tok // This is to avoid the l.readChar() call before this functions return
		} else if utils.IsNumeric(l.char) {
			tok.Type = token.Number
			literal := l.readNumber()
			tok.Literal = literal
			return tok // This is to avoid the l.readChar() call before this functions return
		} else {
			tok = token.MakeToken(token.Illegal, l.char)
		}
	}

	l.readChar()
	return tok
}
