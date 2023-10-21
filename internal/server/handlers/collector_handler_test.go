package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type Metric struct {
	ID    string  `json:"id"`              // Name of metric
	MType string  `json:"type"`            // Gauge or Counter
	Delta int64   `json:"delta,omitempty"` // Value if metric is a counter
	Value float64 `json:"value,omitempty"` // Value if metric is a gauge
}

func TestCollectorHandler_ServeHTTP(t *testing.T) {
	type args struct {
		path   string
		method string
		body   Metric // JSON of metric
	}

	tests := []struct {
		name         string
		args         args
		wantCode     int
		wantErrBody  string
		wantSuccBody Metric
	}{
		{
			name: `Negative #1 (Without body)`,
			args: args{
				path:   `/update/`,
				method: http.MethodPost,
			},
			wantCode:    400,
			wantErrBody: `type not found`,
		},
		{
			name: `Negative #2 (Invalid type)`,
			args: args{
				path:   `/update/`,
				method: http.MethodPost,
				body: Metric{
					ID:    "test",
					MType: "test",
				},
			},
			wantCode:    400,
			wantErrBody: `type not found`,
		},
		{
			name: `Negative #3 (Counter fail)`,
			args: args{
				path:   `/update/`,
				method: http.MethodPost,
				body: Metric{
					ID:    "test",
					MType: "counter",
					Value: 1.1,
				},
			},
			wantCode:    400,
			wantErrBody: `counter value not found`,
		},
		{
			name: `Negative #4 (Gauge fail)`,
			args: args{
				path:   `/update/`,
				method: http.MethodPost,
				body: Metric{
					ID:    "test",
					MType: "gauge",
					Delta: 1,
				},
			},
			wantCode:    400,
			wantErrBody: `gauge value not found`,
		},
		{
			name: `Positive #1 (Counter success)`,
			args: args{
				path:   `/update/`,
				method: http.MethodPost,
				body: Metric{
					ID:    "test",
					MType: "counter",
					Delta: 1,
				},
			},
			wantCode: 200,
			wantSuccBody: Metric{
				ID:    "test",
				MType: "counter",
				Delta: 1,
			},
		},
		{
			name: `Positive #2 (Gauge success)`,
			args: args{
				path:   `/update/`,
				method: http.MethodPost,
				body: Metric{
					ID:    "test",
					MType: "gauge",
					Value: 1.2,
				},
			},
			wantCode: 200,
			wantSuccBody: Metric{
				ID:    "test",
				MType: "gauge",
				Value: 1.2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			g := r.Group(`/update`)
			s := repository.CreateRepository()

			RegisterCollectorHandler(g, s)

			b, _ := json.Marshal(tt.args.body)
			req := httptest.NewRequest(tt.args.method, tt.args.path, strings.NewReader(string(b)))
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
			if tt.wantErrBody != "" {
				assert.Equal(t, tt.wantErrBody, strings.TrimSuffix(rec.Body.String(), "\n"))
				return
			}
			assert.Equal(t, tt.wantSuccBody, tt.args.body)
		})
	}
}
