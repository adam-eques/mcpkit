package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adam-eques/mcpkit/jsonrpc"
	"github.com/adam-eques/mcpkit/mcp"
)

// handleNotification processes an inbound notification. Notifications never
// produce a response; unknown ones are ignored per the JSON-RPC spec.
func (s *Server) handleNotification(_ context.Context, req *jsonrpc.Request) {
	switch req.Method {
	case mcp.NotificationInitialized:
		s.log.Debug("client completed initialization")
	case mcp.NotificationCancelled:
		var p mcp.CancelledParams
		if err := json.Unmarshal(nonNull(req.Params), &p); err != nil {
			s.log.Warn("malformed cancellation", "err", err)
			return
		}
		id := stringifyID(p.RequestID)
		s.log.Debug("cancellation requested", "id", id, "reason", p.Reason)
		s.cancelInflight(id)
	default:
		s.log.Debug("ignoring notification", "method", req.Method)
	}
}

// stringifyID renders a JSON-RPC id (number or string) the same way the inflight
// map keys it, so cancellations match the request they target.
func stringifyID(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%d", int64(t))
	case json.Number:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
