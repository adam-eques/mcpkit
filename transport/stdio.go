package transport

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
)

// MaxFrameBytes bounds a single inbound frame to protect the server from a peer
// that never emits a newline.
const MaxFrameBytes = 16 << 20 // 16 MiB

// ErrFrameTooLarge is returned when an inbound frame exceeds MaxFrameBytes.
var ErrFrameTooLarge = errors.New("transport: frame exceeds maximum size")

// Stdio is a newline-delimited JSON-RPC transport over an io.Reader/io.Writer
// pair, typically os.Stdin and os.Stdout. It is the default MCP transport.
//
// A background goroutine performs the blocking read so that Receive honours
// context cancellation, which the standard bufio reader cannot do on its own.
type Stdio struct {
	w      io.Writer
	closer io.Closer
	writeM sync.Mutex

	frames chan frame
	done   chan struct{}
	closeO sync.Once
}

type frame struct {
	data []byte
	err  error
}

var newline = []byte{'\n'}

// NewStdio wraps r and w in a Stdio transport and starts its read loop. When r
// also implements io.Closer it is closed by Close.
func NewStdio(r io.Reader, w io.Writer) *Stdio {
	s := &Stdio{
		w:      w,
		frames: make(chan frame),
		done:   make(chan struct{}),
	}
	if c, ok := r.(io.Closer); ok {
		s.closer = c
	}
	go s.readLoop(bufio.NewReaderSize(r, 64*1024))
	return s
}

func (s *Stdio) readLoop(br *bufio.Reader) {
	for {
		data, err := readFrame(br)
		select {
		case s.frames <- frame{data: data, err: err}:
		case <-s.done:
			return
		}
		if err != nil {
			return
		}
	}
}

// readFrame reads a single newline-terminated frame, skipping blank lines and
// stripping the trailing CR/LF.
func readFrame(br *bufio.Reader) ([]byte, error) {
	var buf []byte
	for {
		chunk, err := br.ReadSlice('\n')
		buf = append(buf, chunk...)
		if len(buf) > MaxFrameBytes {
			return nil, ErrFrameTooLarge
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		if err != nil {
			line := bytes.TrimRight(buf, "\r\n")
			if len(line) == 0 {
				return nil, err
			}
			return line, nil
		}
		line := bytes.TrimRight(buf, "\r\n")
		if len(line) == 0 {
			buf = buf[:0]
			continue // skip blank keep-alive lines
		}
		return line, nil
	}
}

// Receive implements Transport.
func (s *Stdio) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.done:
		return nil, io.EOF
	case f := <-s.frames:
		return f.data, f.err
	}
}

// Send implements Transport. Writes are serialised and terminated by a newline.
func (s *Stdio) Send(_ context.Context, frame []byte) error {
	if bytes.IndexByte(frame, '\n') >= 0 {
		return errors.New("transport: frame contains an embedded newline")
	}
	s.writeM.Lock()
	defer s.writeM.Unlock()
	if _, err := s.w.Write(frame); err != nil {
		return err
	}
	if _, err := s.w.Write(newline); err != nil {
		return err
	}
	if f, ok := s.w.(interface{ Flush() error }); ok {
		return f.Flush()
	}
	return nil
}

// Close implements Transport.
func (s *Stdio) Close() error {
	var err error
	s.closeO.Do(func() {
		close(s.done)
		if s.closer != nil {
			err = s.closer.Close()
		}
	})
	return err
}
