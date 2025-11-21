package workflow

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenIdent TokenType = iota
	TokenString
	TokenNumber
	TokenBool
	TokenNull
	TokenLParen
	TokenRParen
	TokenComma
	TokenDot
	TokenEq
	TokenNeq
	TokenAnd
	TokenOr
	TokenNot
	TokenLt
	TokenGt
	TokenLte
	TokenGte
	TokenEOF
)

type Token struct {
	Type  TokenType
	Value string
}

func tokenize(expr string) ([]Token, error) {
	var tokens []Token

	r := []rune(expr)
	i := 0

	for i < len(r) {
		// Skip whitespace.
		if unicode.IsSpace(r[i]) {
			i++
			continue
		}

		// Handle two-character operators.
		if i+1 < len(r) {
			two := string(r[i : i+2])
			switch two {
			case "==":
				tokens = append(tokens, Token{TokenEq, "=="})
				i += 2
				continue

			case "!=":
				tokens = append(tokens, Token{TokenNeq, "!="})
				i += 2
				continue

			case "&&":
				tokens = append(tokens, Token{TokenAnd, "&&"})
				i += 2
				continue

			case "||":
				tokens = append(tokens, Token{TokenOr, "||"})
				i += 2
				continue

			case "<=":
				tokens = append(tokens, Token{TokenLte, "<="})
				i += 2
				continue

			case ">=":
				tokens = append(tokens, Token{TokenGte, ">="})
				i += 2
				continue
			}
		}

		// Handle single-character tokens
		switch r[i] {
		case '(':
			tokens = append(tokens, Token{TokenLParen, "("})
			i++
			continue

		case ')':
			tokens = append(tokens, Token{TokenRParen, ")"})
			i++
			continue

		case ',':
			tokens = append(tokens, Token{TokenComma, ","})
			i++
			continue

		case '!':
			tokens = append(tokens, Token{TokenNot, "!"})
			i++
			continue

		case '<':
			tokens = append(tokens, Token{TokenLt, "<"})
			i++
			continue

		case '>':
			tokens = append(tokens, Token{TokenGt, ">"})
			i++
			continue
		}

		// Handle string literal.
		if r[i] == '\'' {
			i++
			start := i

			for i < len(r) && r[i] != '\'' {
				i++
			}

			if i >= len(r) {
				return nil, fmt.Errorf("unterminated string")
			}

			tokens = append(tokens, Token{TokenString, string(r[start:i])})
			i++
			continue
		}

		// Handle number.
		if unicode.IsDigit(r[i]) || (r[i] == '-' && i+1 < len(r) && unicode.IsDigit(r[i+1])) {
			start := i

			if r[i] == '-' {
				i++
			}

			for i < len(r) && (unicode.IsDigit(r[i]) || r[i] == '.') {
				i++
			}

			tokens = append(tokens, Token{TokenNumber, string(r[start:i])})
			continue
		}

		// Handle identifiers (including dot-separated paths).
		if unicode.IsLetter(r[i]) || r[i] == '_' {
			start := i

			for i < len(r) && (unicode.IsLetter(r[i]) || unicode.IsDigit(r[i]) || r[i] == '_' || r[i] == '.') {
				i++
			}

			value := string(r[start:i])

			// Check for keywords.
			switch strings.ToLower(value) {
			case "true":
				tokens = append(tokens, Token{TokenBool, "true"})
			case "false":
				tokens = append(tokens, Token{TokenBool, "false"})
			case "null":
				tokens = append(tokens, Token{TokenNull, "null"})
			default:
				tokens = append(tokens, Token{TokenIdent, value})
			}

			continue
		}

		return nil, fmt.Errorf("unexpected character: %c", r[i])
	}

	tokens = append(tokens, Token{TokenEOF, ""})

	return tokens, nil
}

// Node types for AST.
type Node interface {
	node()
}

type LiteralNode struct {
	Value any
}

func (LiteralNode) node() {}

type ContextAccessNode struct {
	Path string
}

func (ContextAccessNode) node() {}

type BinaryOpNode struct {
	Op    string
	Left  Node
	Right Node
}

func (BinaryOpNode) node() {}

type UnaryOpNode struct {
	Op      string
	Operand Node
}

func (UnaryOpNode) node() {}

type FunctionCallNode struct {
	Name string
	Args []Node
}

func (FunctionCallNode) node() {}

// Parser.
type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{TokenEOF, ""}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() Token {
	t := p.current()
	p.pos++
	return t
}

func (p *parser) parse() (Node, error) {
	return p.parseOr()
}

func (p *parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: "||", Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseAnd() (Node, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenAnd {
		p.advance()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: "&&", Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseEquality() (Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenEq || p.current().Type == TokenNeq {
		op := p.advance().Value
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseComparison() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenLt || p.current().Type == TokenGt || p.current().Type == TokenLte || p.current().Type == TokenGte {
		op := p.advance().Value
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: op, Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseUnary() (Node, error) {
	if p.current().Type == TokenNot {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryOpNode{Op: "!", Operand: operand}, nil
	}

	return p.parsePrimary()
}

func (p *parser) parsePrimary() (Node, error) {
	t := p.current()

	switch t.Type {
	case TokenString:
		p.advance()
		return &LiteralNode{Value: t.Value}, nil

	case TokenNumber:
		p.advance()
		var num float64
		fmt.Sscanf(t.Value, "%f", &num)
		return &LiteralNode{Value: num}, nil

	case TokenBool:
		p.advance()
		return &LiteralNode{Value: t.Value == "true"}, nil

	case TokenNull:
		p.advance()
		return &LiteralNode{Value: nil}, nil

	case TokenIdent:
		p.advance()
		// Check if it's a function call.
		if p.current().Type == TokenLParen {
			return p.parseFunctionCall(t.Value)
		}
		// Otherwise, it's a context access.
		return &ContextAccessNode{Path: t.Value}, nil

	case TokenLParen:
		p.advance()
		node, err := p.parse()
		if err != nil {
			return nil, err
		}
		if p.current().Type != TokenRParen {
			return nil, fmt.Errorf("expected closing paren")
		}
		p.advance()
		return node, nil
	}

	return nil, fmt.Errorf("unexpected token: %v", t)
}

func (p *parser) parseFunctionCall(name string) (Node, error) {
	p.advance() // consume '('.

	var args []Node
	for p.current().Type != TokenRParen {
		arg, err := p.parse()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if p.current().Type == TokenComma {
			p.advance()
		}
	}

	p.advance() // consume ')'.

	return &FunctionCallNode{Name: name, Args: args}, nil
}
