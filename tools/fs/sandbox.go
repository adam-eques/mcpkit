// Package fs implements sandboxed filesystem tools: fs_read, fs_write and
// fs_list. Every path is resolved against a fixed root directory and rejected if
// it would escape that root, so a model cannot read or write arbitrary files.
package fs

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Sandbox confines file operations to Root.
type Sandbox struct {
	Root     string
	ReadOnly bool
	MaxBytes int64
}

// New returns a Sandbox rooted at an absolute, cleaned copy of root.
func New(root string, readOnly bool) (*Sandbox, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Sandbox{Root: filepath.Clean(abs), ReadOnly: readOnly, MaxBytes: 5 << 20}, nil
}

// Resolve turns a caller-supplied relative path into an absolute path inside the
// sandbox, rejecting any attempt to traverse outside Root.
func (s *Sandbox) Resolve(rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute paths are not permitted")
	}
	clean := filepath.Clean(filepath.Join(s.Root, rel))
	if clean != s.Root && !strings.HasPrefix(clean, s.Root+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes the sandbox", rel)
	}
	return clean, nil
}

// Rel returns a path relative to Root for display.
func (s *Sandbox) Rel(abs string) string {
	if rel, err := filepath.Rel(s.Root, abs); err == nil {
		return filepath.ToSlash(rel)
	}
	return abs
}
