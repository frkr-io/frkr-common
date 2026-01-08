package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBrokerWriter(t *testing.T) {
	t.Run("creates writer from connection string", func(t *testing.T) {
		cfg := &Config{
			BrokerURL: "localhost:9092",
		}

		writer := NewBrokerWriter(cfg)
		assert.NotNil(t, writer)
		assert.Equal(t, "localhost:9092", writer.Addr.String())
	})

	t.Run("creates writer from individual components", func(t *testing.T) {
		cfg := &Config{
			BrokerHost: "localhost",
			BrokerPort: "9092",
		}

		writer := NewBrokerWriter(cfg)
		assert.NotNil(t, writer)
		assert.Equal(t, "localhost:9092", writer.Addr.String())
	})

	t.Run("uses default port when BrokerPort not specified", func(t *testing.T) {
		cfg := &Config{
			BrokerHost: "localhost",
		}

		writer := NewBrokerWriter(cfg)
		assert.NotNil(t, writer)
		assert.Equal(t, "localhost:9092", writer.Addr.String())
	})

	t.Run("uses default host and port when neither specified", func(t *testing.T) {
		cfg := &Config{}

		writer := NewBrokerWriter(cfg)
		assert.NotNil(t, writer)
		assert.Equal(t, "localhost:9092", writer.Addr.String())
	})

	t.Run("connection string takes precedence over components", func(t *testing.T) {
		cfg := &Config{
			BrokerURL:  "priority:9092",
			BrokerHost: "ignored",
			BrokerPort: "9999",
		}

		writer := NewBrokerWriter(cfg)
		assert.NotNil(t, writer)
		assert.Equal(t, "priority:9092", writer.Addr.String())
	})
}
