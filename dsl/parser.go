package dsl

import (
	"fmt"
	"strconv"
)

type Spec struct {
	Name         string
	Declarations []Declaration
}

type Declaration interface {
	isDeclaration()
}

type TypeDeclaration struct {
	Identifier        string
	FieldDeclarations []FieldDeclaration
}

func (t TypeDeclaration) isDeclaration() {}

type FieldDeclaration struct {
	Identifier string
	Type       TypeExpression
	Optional   bool
	Nullable   bool
}

type TypeExpression interface {
	isTypeExpression()
}

type SimpleTypeExpression struct {
	Name string
}

func (s SimpleTypeExpression) isTypeExpression() {}

type ArrayTypeExpression struct {
	ElementType TypeExpression
}

func (a ArrayTypeExpression) isTypeExpression() {}

type EndpointDeclaration struct {
	Name   string
	Method string
	Path   string
	Verb   string
	Body   []EndpointFieldDeclaration
}

func (e EndpointDeclaration) isDeclaration() {}

type EndpointFieldDeclaration interface {
	isEndpointField()
}

type ParamsDeclaration struct {
	Fields []FieldDeclaration
}

func (p ParamsDeclaration) isEndpointField() {}

type QueryDeclaration struct {
	Fields []FieldDeclaration
}

func (q QueryDeclaration) isEndpointField() {}

type BodyDeclaration struct {
	Type     SimpleTypeExpression
	Optional bool
}

func (b BodyDeclaration) isEndpointField() {}

type ResponseDeclaration struct {
	Code int
	Type TypeExpression
}

func (r ResponseDeclaration) isEndpointField() {}

type Parser struct {
	input  string
	tokens []Lexeme
	pos    int
}

func (p *Parser) readToken() Lexeme {
	if p.pos >= len(p.tokens) {
		return Lexeme{Type: TokenEOF, Pos: p.pos}
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) peekToken() Lexeme {
	if p.pos >= len(p.tokens) {
		return Lexeme{Type: TokenEOF, Pos: p.pos}
	}
	return p.tokens[p.pos]
}

func (p *Parser) consumeToken() {
	p.pos++
}

func (p *Parser) match(expected Token) error {
	if p.pos >= len(p.tokens) {
		return fmt.Errorf("unexpected end of input, expected %s", expected.String())
	}
	tok := p.readToken()
	if tok.Type != expected {
		return fmt.Errorf("unexpected token %s at position %d, expected %s", tok.String(), tok.Pos, expected.String())
	}
	return nil
}

func (p *Parser) parseFieldDeclarations() ([]FieldDeclaration, error) {
	fieldDecls := []FieldDeclaration{}
	for {
		next := p.peekToken()
		if next.Type != TokenIdentifier {
			break
		}

		fieldNameToken := p.readToken()
		optional := false

		if p.peekToken().Type == TokenQuestionMark {
			p.consumeToken()
			optional = true
		}

		if err := p.match(TokenColon); err != nil {
			return nil, err
		}

		typeToken := p.readToken()
		var typeExpr TypeExpression

		switch typeToken.Type {
		case TokenIdentifier:
			typeExpr = SimpleTypeExpression{Name: typeToken.Value}
		case TokenOpenBracket:
			elemTypeToken := p.readToken()
			if elemTypeToken.Type != TokenIdentifier {
				return nil, fmt.Errorf("unexpected token %s at position %d, expected type identifier", elemTypeToken.String(), elemTypeToken.Pos)
			}
			if err := p.match(TokenCloseBracket); err != nil {
				return nil, err
			}
			typeExpr = ArrayTypeExpression{
				ElementType: SimpleTypeExpression{Name: elemTypeToken.Value},
			}
		case TokenStringType:
			typeExpr = SimpleTypeExpression{Name: "string"}
		case TokenIntType:
			typeExpr = SimpleTypeExpression{Name: "integer"}
		case TokenBoolType:
			typeExpr = SimpleTypeExpression{Name: "boolean"}
		case TokenFloatType:
			typeExpr = SimpleTypeExpression{Name: "float"}
		default:
			return nil, fmt.Errorf("unexpected token %s at position %d, expected type", typeToken.String(), typeToken.Pos)
		}

		nullable := false
		if p.peekToken().Type == TokenQuestionMark {
			p.consumeToken()
			nullable = true
		}

		fieldDecls = append(fieldDecls, FieldDeclaration{
			Identifier: fieldNameToken.Value,
			Type:       typeExpr,
			Optional:   optional,
			Nullable:   nullable,
		})
	}
	return fieldDecls, nil
}

func (p *Parser) parseTypeDeclarations() ([]TypeDeclaration, error) {
	token := p.peekToken()
	if token.Type != TokenType {
		return nil, fmt.Errorf("unexpected token %s at position %d, expected 'type'", token.String(), token.Pos)
	}

	typeDecls := []TypeDeclaration{}
	for {
		if err := p.match(TokenType); err != nil {
			return nil, err
		}

		typeNameToken := p.readToken()
		if typeNameToken.Type != TokenIdentifier {
			return nil, fmt.Errorf("unexpected token %s at position %d, expected type name", typeNameToken.String(), typeNameToken.Pos)
		}

		if err := p.match(TokenOpenBrace); err != nil {
			return nil, err
		}

		fieldDecls, err := p.parseFieldDeclarations()
		if err != nil {
			return nil, err
		}

		if err := p.match(TokenCloseBrace); err != nil {
			return nil, err
		}

		typeDecls = append(typeDecls, TypeDeclaration{
			Identifier:        typeNameToken.Value,
			FieldDeclarations: fieldDecls,
		})

		next := p.peekToken()
		if next.Type != TokenType {
			break
		}
	}
	return typeDecls, nil
}

func (p *Parser) parseResponseDeclarations() ([]ResponseDeclaration, error) {
	if err := p.match(TokenResponses); err != nil {
		return nil, err
	}
	if err := p.match(TokenOpenBrace); err != nil {
		return nil, err
	}
	responses := []ResponseDeclaration{}
	for {
		codeToken := p.peekToken()
		if codeToken.Type != TokenNumberLiteral {
			break
		}
		p.consumeToken()

		typeToken := p.readToken()
		if typeToken.Type != TokenIdentifier && !typeToken.Type.IsType() {
			return nil, fmt.Errorf("unexpected token %s at position %d, expected type identifier", typeToken.String(), typeToken.Pos)
		}

		code, err := strconv.Atoi(codeToken.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid response code %s at position %d", codeToken.Value, codeToken.Pos)
		}
		responses = append(responses, ResponseDeclaration{
			Code: code,
			Type: SimpleTypeExpression{Name: typeToken.Value},
		})
	}
	if err := p.match(TokenCloseBrace); err != nil {
		return nil, err
	}
	return responses, nil
}

func (p *Parser) parseEndpointBody() ([]EndpointFieldDeclaration, error) {
	fields := []EndpointFieldDeclaration{}
	token := p.peekToken()

	if token.Type == TokenParams {
		p.consumeToken()
		if err := p.match(TokenOpenBrace); err != nil {
			return nil, err
		}
		fieldDecls, err := p.parseFieldDeclarations()
		if err != nil {
			return nil, err
		}
		if err := p.match(TokenCloseBrace); err != nil {
			return nil, err
		}
		fields = append(fields, ParamsDeclaration{Fields: fieldDecls})
		token = p.peekToken()
	}

	token = p.peekToken()
	if token.Type == TokenQuery {
		p.consumeToken()
		if err := p.match(TokenOpenBrace); err != nil {
			return nil, err
		}
		fieldDecls, err := p.parseFieldDeclarations()
		if err != nil {
			return nil, err
		}
		if err := p.match(TokenCloseBrace); err != nil {
			return nil, err
		}
		fields = append(fields, QueryDeclaration{Fields: fieldDecls})
	}

	token = p.peekToken()
	if token.Type == TokenBody {
		p.consumeToken()
		typeToken := p.readToken()
		if typeToken.Type != TokenIdentifier {
			return nil, fmt.Errorf("unexpected token %s at position %d, expected type identifier", typeToken.String(), typeToken.Pos)
		}

		optional := false
		if p.peekToken().Type == TokenQuestionMark {
			p.consumeToken()
			optional = true
		}

		fields = append(fields, BodyDeclaration{
			Type:     SimpleTypeExpression{Name: typeToken.Value},
			Optional: optional,
		})
	}

	responses, err := p.parseResponseDeclarations()
	if err != nil {
		return nil, err
	}

	for _, resp := range responses {
		fields = append(fields, resp)
	}

	return fields, nil
}

func (p *Parser) parseEndpointDeclarations() ([]EndpointDeclaration, error) {
	endpointDecls := []EndpointDeclaration{}
	token := p.peekToken()
	for token.Type == TokenEndpoint {
		p.consumeToken()
		methodToken := p.readToken()
		if !methodToken.Type.IsHTTPMethod() {
			return nil, fmt.Errorf("unexpected token %s at position %d, expected HTTP method", methodToken.String(), methodToken.Pos)
		}

		pathToken := p.readToken()
		if pathToken.Type != TokenPath {
			return nil, fmt.Errorf("unexpected token %s at position %d, expected path", pathToken.String(), pathToken.Pos)
		}

		endpointNameToken := p.readToken()
		if endpointNameToken.Type != TokenIdentifier {
			return nil, fmt.Errorf("unexpected token %s at position %d, expected endpoint name", endpointNameToken.String(), endpointNameToken.Pos)
		}

		if err := p.match(TokenOpenBrace); err != nil {
			return nil, err
		}
		body, err := p.parseEndpointBody()
		if err != nil {
			return nil, err
		}
		if err := p.match(TokenCloseBrace); err != nil {
			return nil, err
		}

		endpointDecls = append(endpointDecls, EndpointDeclaration{
			Name:   endpointNameToken.Value,
			Method: methodToken.Value,
			Path:   pathToken.Value,
			Body:   body,
		})
		token = p.peekToken()
	}
	return endpointDecls, nil
}

func (p *Parser) parseSpec() (*Spec, error) {
	if err := p.match(TokenAPI); err != nil {
		return nil, err
	}
	nameToken := p.readToken()
	if nameToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("unexpected token %s at position %d, expected identifier", nameToken.String(), nameToken.Pos)
	}

	decs := []Declaration{}
	next := p.peekToken()
	if next.Type == TokenType {
		decl, err := p.parseTypeDeclarations()
		if err != nil {
			return nil, err
		}
		for _, td := range decl {
			decs = append(decs, td)
		}
	}

	endpoints, err := p.parseEndpointDeclarations()
	if err != nil {
		return nil, err
	}
	for _, ed := range endpoints {
		decs = append(decs, ed)
	}

	return &Spec{
		Name:         nameToken.Value,
		Declarations: decs,
	}, nil
}

func (p *Parser) Parse() (*Spec, error) {
	spec, err := p.parseSpec()
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func NewParserFromString(input string) *Parser {
	lexer := NewLexer(input)
	tokens := lexer.Tokenize()
	return NewParser(input, tokens)
}

func NewParser(input string, tokens []Lexeme) *Parser {
	return &Parser{
		input:  input,
		tokens: tokens,
		pos:    0,
	}
}
