package mcp

// PingResult is the empty body returned by ping.
type PingResult struct{}

// ProgressParams report incremental progress for a long-running request.
type ProgressParams struct {
	ProgressToken any     `json:"progressToken"`
	Progress      float64 `json:"progress"`
	Total         float64 `json:"total,omitempty"`
	Message       string  `json:"message,omitempty"`
}

// CancelledParams request cancellation of an in-flight request.
type CancelledParams struct {
	RequestID any    `json:"requestId"`
	Reason    string `json:"reason,omitempty"`
}

// LoggingLevel is a syslog-style severity understood by logging/setLevel.
type LoggingLevel string

const (
	LogDebug     LoggingLevel = "debug"
	LogInfo      LoggingLevel = "info"
	LogNotice    LoggingLevel = "notice"
	LogWarning   LoggingLevel = "warning"
	LogError     LoggingLevel = "error"
	LogCritical  LoggingLevel = "critical"
	LogAlert     LoggingLevel = "alert"
	LogEmergency LoggingLevel = "emergency"
)

// SetLevelParams are the parameters of logging/setLevel.
type SetLevelParams struct {
	Level LoggingLevel `json:"level"`
}

// LogMessageParams carry a server-emitted log record.
type LogMessageParams struct {
	Level  LoggingLevel `json:"level"`
	Logger string       `json:"logger,omitempty"`
	Data   any          `json:"data"`
}
