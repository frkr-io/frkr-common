package gateway

import (
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
)

// CreateTopicIfNotExists creates a Kafka topic if it doesn't already exist
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
