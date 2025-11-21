// Package mcp defines the Model Context Protocol message types exchanged over a
// JSON-RPC transport: the initialize handshake, capability negotiation, and the
// tools, resources and prompts primitives.
package mcp

import "slices"

// ProtocolVersion is the newest MCP revision this library implements.
const ProtocolVersion = "2025-06-18"

// SupportedProtocolVersions lists the revisions the server can negotiate, newest
// first. During initialize the server selects the client's requested version
// when supported and otherwise falls back to ProtocolVersion.
var SupportedProtocolVersions = []string{
	"2025-06-18",
	"2025-03-26",
	"2024-11-05",
}

// NegotiateVersion returns the version to use given the client's request. When
// the requested version is supported it is echoed back; otherwise the server's
// preferred version is returned so the client can decide whether to proceed.
func NegotiateVersion(requested string) string {
	if slices.Contains(SupportedProtocolVersions, requested) {
		return requested
	}
	return ProtocolVersion
}
