package transport

import (
	"context"
	"io"
	"sync"
)

// Pipe returns a pair of connected in-memory transports. A frame sent on one end
// is received on the other. It is used by tests and examples to drive a server
// without spawning a process.
func Pipe() (client, server Transport) {
	a := make(chan []byte, 16)
	b := make(chan []byte, 16)
	done := make(chan struct{})
	var once sync.Once
	closeFn := func() { once.Do(func() { close(done) }) }
	return &pipeEnd{send: a, recv: b, done: done, closeFn: closeFn},
		&pipeEnd{send: b, recv: a, done: done, closeFn: closeFn}
}

type pipeEnd struct {
	send    chan []byte
	recv    chan []byte
	done    chan struct{}
	closeFn func()
}

func (p *pipeEnd) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.done:
		return nil, io.EOF
	case data := <-p.recv:
		return data, nil
	}
}

func (p *pipeEnd) Send(ctx context.Context, frame []byte) error {
	cp := append([]byte(nil), frame...)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return io.ErrClosedPipe
	case p.send <- cp:
		return nil
	}
}

func (p *pipeEnd) Close() error {
	p.closeFn()
	return nil
}
