// Package calc implements a safe arithmetic expression evaluator exposed as the
// "calculate" tool. It uses a hand-written tokenizer and recursive-descent
// parser supporting the usual operators, parentheses, unary signs, a set of
// mathematical functions and the constants pi and e. No user input ever reaches
// a shell or the Go evaluator, so the tool is safe to expose to a model.
package calc

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// Eval parses and evaluates a mathematical expression.
func Eval(expr string) (float64, error) {
	toks, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	p := &parser{toks: toks}
	v, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	if p.pos != len(p.toks) {
		return 0, fmt.Errorf("unexpected token %q", p.toks[p.pos].text)
	}
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, errors.New("result is not a finite number")
	}
	return v, nil
}

type tokenKind int

const (
	tokNumber tokenKind = iota
	tokIdent
	tokOp
	tokLParen
	tokRParen
)

type token struct {
	kind tokenKind
	text string
	num  float64
}

func tokenize(s string) ([]token, error) {
	var toks []token
	runes := []rune(s)
	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case unicode.IsSpace(r):
			i++
		case r == '(':
			toks = append(toks, token{kind: tokLParen, text: "("})
			i++
		case r == ')':
			toks = append(toks, token{kind: tokRParen, text: ")"})
			i++
		case strings.ContainsRune("+-*/%^", r):
			toks = append(toks, token{kind: tokOp, text: string(r)})
			i++
		case r == '.' || unicode.IsDigit(r):
			j := i
			for j < len(runes) && (unicode.IsDigit(runes[j]) || runes[j] == '.' ||
				runes[j] == 'e' || runes[j] == 'E' ||
				((runes[j] == '+' || runes[j] == '-') && j > i && (runes[j-1] == 'e' || runes[j-1] == 'E'))) {
				j++
			}
			lit := string(runes[i:j])
			v, err := strconv.ParseFloat(lit, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number %q", lit)
			}
			toks = append(toks, token{kind: tokNumber, text: lit, num: v})
			i = j
		case unicode.IsLetter(r) || r == '_':
			j := i
			for j < len(runes) && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j]) || runes[j] == '_') {
				j++
			}
			toks = append(toks, token{kind: tokIdent, text: strings.ToLower(string(runes[i:j]))})
			i = j
		default:
			return nil, fmt.Errorf("unexpected character %q", string(r))
		}
	}
	return toks, nil
}
