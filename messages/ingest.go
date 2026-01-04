package messages

// IngestRequest represents a request to ingest a mirrored API request
type IngestRequest struct {
	StreamID string         `json:"stream_id"`
	Request  MirroredRequest `json:"request"`
}

// MirroredRequest represents a mirrored HTTP request
type MirroredRequest struct {
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Query      map[string]string `json:"query"`
	TimestampNS int64            `json:"timestamp_ns"`
	RequestID   string            `json:"request_id"`
}

// IngestResponse represents the response from the ingest gateway
type IngestResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

