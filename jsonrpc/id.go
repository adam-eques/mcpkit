package jsonrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
)

// ID is a JSON-RPC 2.0 request identifier. Per the specification an identifier
// may be a string or an integer; a request that omits the identifier entirely is
// a notification. ID preserves the original representation so that a response
// echoes back the exact identifier the peer supplied.
type ID struct {
	num   int64
	str   string
	isStr bool
	valid bool
}

// Int64ID returns an integer identifier.
func Int64ID(n int64) ID { return ID{num: n, valid: true} }

// StringID returns a string identifier.
func StringID(s string) ID { return ID{str: s, isStr: true, valid: true} }

// IsValid reports whether the identifier carries a value (as opposed to null).
func (id ID) IsValid() bool { return id.valid }

// IsString reports whether the identifier is a string.
func (id ID) IsString() bool { return id.isStr }

// String renders the identifier for logging and map keys.
func (id ID) String() string {
	switch {
	case !id.valid:
		return "null"
	case id.isStr:
		return id.str
	default:
		return strconv.FormatInt(id.num, 10)
	}
}

// MarshalJSON implements json.Marshaler.
func (id ID) MarshalJSON() ([]byte, error) {
	switch {
	case !id.valid:
		return []byte("null"), nil
	case id.isStr:
		return json.Marshal(id.str)
	default:
		return strconv.AppendInt(nil, id.num, 10), nil
	}
}

// UnmarshalJSON implements json.Unmarshaler. Fractional numbers are rejected
// because JSON-RPC identifiers must not contain a fractional part.
func (id *ID) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*id = ID{}
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*id = ID{str: s, isStr: true, valid: true}
		return nil
	}
	if bytes.ContainsAny(data, ".eE") {
		return errors.New("jsonrpc: identifier must be an integer or a string")
	}
	n, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*id = ID{num: n, valid: true}
	return nil
}
