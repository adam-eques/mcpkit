package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/adam-eques/mcpkit/jsonrpc"
	"github.com/adam-eques/mcpkit/transport"
)

// Serve reads frames from t and dispatches them until the peer closes the
// stream, ctx is cancelled, or a fatal transport error occurs. Requests are
// handled on their own goroutines subject to the configured concurrency limit,
// and the transport serialises the responses.
func (s *Server) Serve(ctx context.Context, t transport.Transport) error {
	defer t.Close()
	var wg sync.WaitGroup
	for {
		frame, err := t.Receive(ctx)
		if err != nil {
			wg.Wait()
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		if len(frame) == 0 {
			continue
		}
		s.acquire()
		wg.Add(1)
		go func(frame []byte) {
			defer wg.Done()
			defer s.release()
			s.dispatchFrame(ctx, t, frame)
		}(frame)
	}
}

// dispatchFrame handles one frame, wiring per-request cancellation so that a
// notifications/cancelled from the peer aborts the in-flight context.
func (s *Server) dispatchFrame(ctx context.Context, t transport.Transport, frame []byte) {
	reqCtx := ctx
	if id, isRequest := peekID(frame); isRequest {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithCancel(ctx)
		s.registerInflight(id, cancel)
		defer s.clearInflight(id)
	}

	resp, err := s.HandleMessage(reqCtx, frame)
	if err != nil {
		s.log.Error("handle message", "err", err)
		return
	}
	if resp == nil {
		return // notification
	}
	if err := t.Send(ctx, resp); err != nil {
		s.log.Error("send response", "err", err)
	}
}

// peekID extracts the request id without fully decoding the frame. It reports
// isRequest=false for notifications, which have no id.
func peekID(frame []byte) (id string, isRequest bool) {
	var head struct {
		ID *jsonrpc.ID `json:"id"`
	}
	if err := json.Unmarshal(frame, &head); err != nil || head.ID == nil {
		return "", false
	}
	return head.ID.String(), true
}
