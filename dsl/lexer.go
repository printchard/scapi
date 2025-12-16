package dsl

import (
	"fmt"
	"unicode"
)

type Token int

const (
	TokenEOF Token = iota
	TokenType
	TokenOpenBrace
	TokenCloseBrace
	TokenOpenBracket
	TokenCloseBracket
	TokenAPI
	TokenEndpoint
	TokenGetMethod
	TokenPostMethod
	TokenPutMethod
	TokenDeleteMethod
	TokenPatchMethod
	TokenIdentifier
	TokenStringType
	TokenIntType
	TokenBoolType
	TokenFloatType
	TokenColon
	TokenParams
	TokenQuery
	TokenBody
	TokenResponses
	TokenNumberLiteral
	TokenQuestionMark
	TokenPath
)

var tokenNames = map[Token]string{
	TokenEOF:           "EOF",
	TokenType:          "TYPE",
	TokenOpenBrace:     "{",
	TokenCloseBrace:    "}",
	TokenOpenBracket:   "[",
	TokenCloseBracket:  "]",
	TokenAPI:           "API",
	TokenEndpoint:      "ENDPOINT",
	TokenGetMethod:     "GET",
	TokenPostMethod:    "POST",
	TokenPutMethod:     "PUT",
	TokenDeleteMethod:  "DELETE",
	TokenPatchMethod:   "PATCH",
	TokenIdentifier:    "IDENTIFIER",
	TokenStringType:    "STRING",
	TokenIntType:       "INT",
	TokenBoolType:      "BOOL",
	TokenFloatType:     "FLOAT",
	TokenColon:         ":",
	TokenParams:        "PARAMS",
	TokenQuery:         "QUERY",
	TokenBody:          "BODY",
	TokenResponses:     "RESPONSES",
	TokenNumberLiteral: "NUMBER_LITERAL",
	TokenQuestionMark:  "?",
	TokenPath:          "PATH",
}

func (t Token) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

func (t Token) IsHTTPMethod() bool {
	return t == TokenGetMethod || t == TokenPostMethod || t == TokenPutMethod || t == TokenDeleteMethod || t == TokenPatchMethod
}

func (t Token) IsType() bool {
	return t == TokenStringType || t == TokenIntType || t == TokenBoolType || t == TokenFloatType
}

type Lexeme struct {
	Type  Token
	Pos   int
	Value string
}

func (l Lexeme) String() string {
	switch l.Type {
	case TokenIdentifier, TokenPath, TokenNumberLiteral:
		return fmt.Sprintf("%s(%s)", l.Type.String(), l.Value)
	default:
		return l.Type.String()
	}
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
	}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

func isWhitespace(ch byte) bool {
	return unicode.IsSpace(rune(ch))
}

func (l *Lexer) consumeChar() {
	l.pos++
}

func (l *Lexer) readChar() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	return ch
}

func (l *Lexer) peekChar() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func lookupIdent(ident string) Token {
	switch ident {
	case "type":
		return TokenType
	case "api":
		return TokenAPI
	case "endpoint":
		return TokenEndpoint
	case "GET":
		return TokenGetMethod
	case "POST":
		return TokenPostMethod
	case "PUT":
		return TokenPutMethod
	case "DELETE":
		return TokenDeleteMethod
	case "PATCH":
		return TokenPatchMethod
	case "string":
		return TokenStringType
	case "integer":
		return TokenIntType
	case "boolean":
		return TokenBoolType
	case "float":
		return TokenFloatType
	case "params":
		return TokenParams
	case "query":
		return TokenQuery
	case "body":
		return TokenBody
	case "responses":
		return TokenResponses
	default:
		return TokenIdentifier
	}
}

func (l *Lexer) readNumber() string {
	startPos := l.pos - 1
	for isDigit(l.peekChar()) {
		l.consumeChar()
	}
	if l.peekChar() == '.' {
		l.consumeChar()
		for isDigit(l.peekChar()) {
			l.consumeChar()
		}
	}
	return l.input[startPos:l.pos]
}

func (l *Lexer) readPath() string {
	startPos := l.pos - 1
	for !isWhitespace(l.peekChar()) {
		l.consumeChar()
	}
	return l.input[startPos:l.pos]
}

func (l *Lexer) NextToken() Lexeme {
	c := l.readChar()

	for isWhitespace(c) {
		c = l.readChar()
	}

	switch c {
	case 0:
		return Lexeme{Type: TokenEOF, Pos: l.pos}
	case '{':
		return Lexeme{Type: TokenOpenBrace, Pos: l.pos - 1}
	case '}':
		return Lexeme{Type: TokenCloseBrace, Pos: l.pos - 1}
	case ':':
		return Lexeme{Type: TokenColon, Pos: l.pos - 1}
	case '?':
		return Lexeme{Type: TokenQuestionMark, Pos: l.pos - 1}
	case '[':
		return Lexeme{Type: TokenOpenBracket, Pos: l.pos - 1}
	case ']':
		return Lexeme{Type: TokenCloseBracket, Pos: l.pos - 1}
	case '/':
		if l.peekChar() == '/' {
			for c != '\n' && c != 0 {
				c = l.readChar()
			}
			return l.NextToken()
		}
		startPos := l.pos - 1
		path := l.readPath()
		return Lexeme{Type: TokenPath, Pos: startPos, Value: path}
	default:
		if isLetter(c) {
			startPos := l.pos - 1
			for isLetter(l.peekChar()) || isDigit(l.peekChar()) {
				l.consumeChar()
			}
			ident := l.input[startPos:l.pos]
			tokenType := lookupIdent(ident)
			return Lexeme{Type: tokenType, Pos: startPos, Value: ident}
		} else if isDigit(c) {
			startPos := l.pos - 1
			number := l.readNumber()
			return Lexeme{Type: TokenNumberLiteral, Pos: startPos, Value: number}
		}
	}
	return Lexeme{Type: TokenEOF, Pos: l.pos}
}

func (l *Lexer) Tokenize() []Lexeme {
	lexeme := l.NextToken()
	var lexemes []Lexeme
	for lexeme.Type != TokenEOF {
		lexemes = append(lexemes, lexeme)
		lexeme = l.NextToken()
	}
	lexemes = append(lexemes, lexeme)
	return lexemes
}
