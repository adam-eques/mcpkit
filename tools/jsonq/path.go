// Package jsonq implements the "json_query" tool: it evaluates a small dotted
// path expression against a JSON document, e.g. "user.roles[0].name". Supported
// steps are object keys, bracketed keys ["a.b"] and array indices [n].
package jsonq

import (
	"fmt"
	"strconv"
	"strings"
)

// Query evaluates path against the decoded JSON value doc.
func Query(doc any, path string) (any, error) {
	steps, err := parsePath(path)
	if err != nil {
		return nil, err
	}
	cur := doc
	for i, step := range steps {
		next, err := step.apply(cur)
		if err != nil {
			return nil, fmt.Errorf("at %q: %w", pathPrefix(steps, i), err)
		}
		cur = next
	}
	return cur, nil
}

type step interface {
	apply(any) (any, error)
	String() string
}

type keyStep struct{ key string }

func (k keyStep) apply(v any) (any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object, got %T", v)
	}
	val, ok := m[k.key]
	if !ok {
		return nil, fmt.Errorf("key %q not found", k.key)
	}
	return val, nil
}

func (k keyStep) String() string { return k.key }

type indexStep struct{ idx int }

func (s indexStep) apply(v any) (any, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("expected array, got %T", v)
	}
	if s.idx < 0 || s.idx >= len(arr) {
		return nil, fmt.Errorf("index %d out of range (len %d)", s.idx, len(arr))
	}
	return arr[s.idx], nil
}

func (s indexStep) String() string { return "[" + strconv.Itoa(s.idx) + "]" }

func parsePath(path string) ([]step, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "." || path == "$" {
		return nil, nil
	}
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")

	var steps []step
	i := 0
	for i < len(path) {
		switch path[i] {
		case '.':
			i++
		case '[':
			end := strings.IndexByte(path[i:], ']')
			if end < 0 {
				return nil, fmt.Errorf("unclosed bracket in path")
			}
			inner := path[i+1 : i+end]
			i += end + 1
			if len(inner) >= 2 && (inner[0] == '"' || inner[0] == '\'') {
				steps = append(steps, keyStep{key: inner[1 : len(inner)-1]})
			} else {
				n, err := strconv.Atoi(inner)
				if err != nil {
					return nil, fmt.Errorf("invalid array index %q", inner)
				}
				steps = append(steps, indexStep{idx: n})
			}
		default:
			j := i
			for j < len(path) && path[j] != '.' && path[j] != '[' {
				j++
			}
			steps = append(steps, keyStep{key: path[i:j]})
			i = j
		}
	}
	return steps, nil
}

func pathPrefix(steps []step, upto int) string {
	parts := make([]string, 0, upto+1)
	for i := 0; i <= upto && i < len(steps); i++ {
		parts = append(parts, steps[i].String())
	}
	return strings.Join(parts, ".")
}
