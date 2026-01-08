package gateway

import (
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// NewBrokerWriter creates a new broker writer from GatewayBaseConfig
// Supports both connection string (BrokerURL) and individual components
// Connection string takes precedence if both are provided
func NewBrokerWriter(cfg *GatewayBaseConfig) *kafka.Writer {
	var brokerURL string

	// Prefer connection string if provided
	if cfg.BrokerURL != "" {
		brokerURL = cfg.BrokerURL
	} else {
		// Build URL from individual components
		if cfg.BrokerHost == "" {
			brokerURL = "localhost:9092" // Default
		} else {
			port := cfg.BrokerPort
			if port == "" {
				port = "9092" // Default Kafka port
			}
			brokerURL = fmt.Sprintf("%s:%s", cfg.BrokerHost, port)
		}
	}

	return &kafka.Writer{
		Addr:         kafka.TCP(brokerURL),
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
	}
}

// CreateTopicIfNotExists creates a topic if it doesn't already exist
func CreateTopicIfNotExists(brokerURL, topicName string) error {
	conn, err := kafka.Dial("tcp", brokerURL)
	if err != nil {
		return fmt.Errorf("failed to connect to broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "already exists") || strings.Contains(errStr, "TOPIC_ALREADY_EXISTS") {
			return nil
		}
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
