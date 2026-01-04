package messages

// StreamMessage represents a message streamed to CLI clients
type StreamMessage struct {
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Query      map[string]string `json:"query"`
	TimestampNS int64            `json:"timestamp_ns"`
	RequestID   string            `json:"request_id"`
}

