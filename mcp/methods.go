package mcp

// Request methods defined by the protocol.
const (
	MethodInitialize             = "initialize"
	MethodPing                   = "ping"
	MethodToolsList              = "tools/list"
	MethodToolsCall              = "tools/call"
	MethodResourcesList          = "resources/list"
	MethodResourcesRead          = "resources/read"
	MethodResourcesTemplatesList = "resources/templates/list"
	MethodPromptsList            = "prompts/list"
	MethodPromptsGet             = "prompts/get"
	MethodLoggingSetLevel        = "logging/setLevel"
)

// Notification methods. Notifications carry no identifier and expect no reply.
const (
	NotificationInitialized      = "notifications/initialized"
	NotificationCancelled        = "notifications/cancelled"
	NotificationProgress         = "notifications/progress"
	NotificationMessage          = "notifications/message"
	NotificationToolsListChanged = "notifications/tools/list_changed"
)
