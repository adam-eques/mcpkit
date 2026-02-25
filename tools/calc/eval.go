package calc

import (
	"fmt"
	"math"
)

type parser struct {
	toks []token
	pos  int
}

func (p *parser) peek() (token, bool) {
	if p.pos < len(p.toks) {
		return p.toks[p.pos], true
	}
	return token{}, false
}

func (p *parser) next() token {
	t := p.toks[p.pos]
	p.pos++
	return t
}

// expr := term (('+' | '-') term)*
func (p *parser) parseExpr() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		t, ok := p.peek()
		if !ok || t.kind != tokOp || (t.text != "+" && t.text != "-") {
			return left, nil
		}
		p.next()
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if t.text == "+" {
			left += right
		} else {
			left -= right
		}
	}
}

// term := unary (('*' | '/' | '%') unary)*
func (p *parser) parseTerm() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		t, ok := p.peek()
		if !ok || t.kind != tokOp || (t.text != "*" && t.text != "/" && t.text != "%") {
			return left, nil
		}
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		switch t.text {
		case "*":
			left *= right
		case "/":
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		case "%":
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			left = math.Mod(left, right)
		}
	}
}

// unary := ('+' | '-') unary | power
//
// Unary sign binds looser than exponentiation, so -2^2 evaluates to -(2^2).
func (p *parser) parseUnary() (float64, error) {
	if t, ok := p.peek(); ok && t.kind == tokOp && (t.text == "+" || t.text == "-") {
		p.next()
		v, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		if t.text == "-" {
			return -v, nil
		}
		return v, nil
	}
	return p.parsePower()
}

// power := primary ('^' unary)?  (right associative; 2^-3 is allowed)
func (p *parser) parsePower() (float64, error) {
	base, err := p.parsePrimary()
	if err != nil {
		return 0, err
	}
	if t, ok := p.peek(); ok && t.kind == tokOp && t.text == "^" {
		p.next()
		exp, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return math.Pow(base, exp), nil
	}
	return base, nil
}

// primary := number | constant | func '(' expr ')' | '(' expr ')'
func (p *parser) parsePrimary() (float64, error) {
	t, ok := p.peek()
	if !ok {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	switch t.kind {
	case tokNumber:
		p.next()
		return t.num, nil
	case tokLParen:
		p.next()
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if end, ok := p.peek(); !ok || end.kind != tokRParen {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.next()
		return v, nil
	case tokIdent:
		return p.parseIdent()
	default:
		return 0, fmt.Errorf("unexpected token %q", t.text)
	}
}

func (p *parser) parseIdent() (float64, error) {
	name := p.next().text
	if c, ok := constants[name]; ok {
		return c, nil
	}
	fn, ok := functions[name]
	if !ok {
		return 0, fmt.Errorf("unknown identifier %q", name)
	}
	if t, ok := p.peek(); !ok || t.kind != tokLParen {
		return 0, fmt.Errorf("function %q must be called with parentheses", name)
	}
	p.next() // consume '('
	arg, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	if end, ok := p.peek(); !ok || end.kind != tokRParen {
		return 0, fmt.Errorf("missing closing parenthesis after %q", name)
	}
	p.next()
	return fn(arg), nil
}

var constants = map[string]float64{
	"pi":  math.Pi,
	"e":   math.E,
	"tau": 2 * math.Pi,
}

var functions = map[string]func(float64) float64{
	"sqrt":  math.Sqrt,
	"abs":   math.Abs,
	"sin":   math.Sin,
	"cos":   math.Cos,
	"tan":   math.Tan,
	"asin":  math.Asin,
	"acos":  math.Acos,
	"atan":  math.Atan,
	"log":   math.Log10,
	"ln":    math.Log,
	"exp":   math.Exp,
	"floor": math.Floor,
	"ceil":  math.Ceil,
	"round": math.Round,
}
