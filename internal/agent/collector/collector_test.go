package collector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCollector_StartTickers(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: `Positive`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CreateCollector()

			go c.StartTickers()
			c.CloseChannel()

			_, ok := <-c.done
			assert.False(t, ok)
		})
	}
}

func TestCollector_CollectMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: `Positive`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CreateCollector()
			go c.StartTickers()

			time.Sleep(time.Duration(3) * time.Second)

			c.CloseChannel()

			// Check metrics
			assert.NotEqual(t, 0, len(c.counter))
			assert.NotEqual(t, 0, len(c.gauge))

			// Check amount of poll count
			assert.Equal(t, int64(1), c.counter[`PollCount`])
		})
	}
}
