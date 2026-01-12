package db

import (
	"fmt"
	"strings"
)

// GenerateTopicName generates a Kafka-compliant topic name from tenant ID and stream name
func GenerateTopicName(tenantID, streamName string) string {
	// Sanitize for topic name (lowercase, replace spaces/special chars with hyphens)
	topic := fmt.Sprintf("stream-%s-%s", 
		strings.ToLower(strings.ReplaceAll(tenantID, "-", "")),
		strings.ToLower(strings.ReplaceAll(streamName, " ", "-")))
	
	// Remove any remaining invalid characters
	topic = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, topic)
	
	// Truncate to 249 chars to be safe (Kafka limit is 255)
	if len(topic) > 249 {
		topic = topic[:249]
	}
	
	return topic
}
