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

// TestAppBodyHandlers tests the new API where we can send body.
func TestAppBodyHandlers(t *testing.T) {
	type args struct {
		path   string
		method string
		body   Metric
	}
	tests := []struct {
		name         string
		args         args
		wantCode     int
		wantErrBody  string
		wantSuccBody Metric
	}{
		{
			name: `Negative #1 (Live not found)`,
			args: args{
				path:   `/live`,
				method: http.MethodPost,
			},
			wantCode:    404,
			wantErrBody: `404 page not found`,
		},
		{
			name: `Negative #2 (Metric without body)`,
			args: args{
				path:   `/update`,
				method: http.MethodPost,
			},
			wantCode:    400,
			wantErrBody: `type is invalid or not found`,
		},
		{
			name: `Negative #3 (Metric with invalid type)`,
			args: args{
				path:   `/update`,
				method: http.MethodPost,
				body: Metric{
					ID:    "test",
					MType: "invalid",
				},
			},
			wantCode:    400,
			wantErrBody: `type is invalid or not found`,
		},
		{
			name: `Negative #4 (Metric counter fail)`,
			args: args{
				path:   `/update`,
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
			name: `Negative #5 (Metric gauge fail)`,
			args: args{
				path:   `/update`,
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
			name: `Negative #1 (Value without body)`,
			args: args{
				path: `/value`,
			},
			wantCode:    404,
			wantErrBody: `404 page not found`,
		},
		{
			name: `Negative #2 (Value unknown name)`,
			args: args{
				path: `/value`,
				body: Metric{
					ID:    "t.e.s.t",
					MType: "counter",
				},
			},
			wantCode:    404,
			wantErrBody: `404 page not found`,
		},
		{
			name: `Positive #1 (Live success)`,
			args: args{
				path:   `/live`,
				method: http.MethodGet,
			},
			wantCode:    200,
			wantErrBody: `Live`,
		},
		{
			name: `Positive #2 (Metric counter success)`,
			args: args{
				path:   `/update`,
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
			name: `Positive #3 (Metric gauge success)`,
			args: args{
				path:   `/update`,
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
			g := r.Group(`/`)
			s := repository.CreateRepository()

			RegisterAppHandler(g, s)

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

// TestAppParamsHandlers tests the old API where we send params.
func TestAppParamsHandlers(t *testing.T) {
	type args struct {
		path   string
		method string
	}
	tests := []struct {
		name        string
		args        args
		wantCode    int
		wantErrBody string
	}{
		{
			name: `Negative #1 (Metric without name)`,
			args: args{
				path:   `/update/counter`,
				method: http.MethodPost,
			},
			wantCode:    404,
			wantErrBody: `name is invalid or not found`,
		},
		{
			name: `Negative #2 (Metric with invalid type)`,
			args: args{
				path:   `/update/invalid/test/1`,
				method: http.MethodPost,
			},
			wantCode:    400,
			wantErrBody: `type is invalid or not found`,
		},
		{
			name: `Negative #3 (Metric without value)`,
			args: args{
				path:   `/update/counter/testCounter`,
				method: http.MethodPost,
			},
			wantCode:    400,
			wantErrBody: `value is invalid or not found`,
		},
		{
			name: `Negative #4 (Metric counter fail)`,
			args: args{
				path:   `/update/counter/test/1.1`,
				method: http.MethodPost,
			},
			wantCode:    400,
			wantErrBody: `counter value not found`,
		},
		{
			name: `Negative #5 (Value unknown name)`,
			args: args{
				path: `/value/counter/tist`,
			},
			wantCode:    404,
			wantErrBody: `404 page not found`,
		},
		{
			name: `Positive #1 (Metric counter success)`,
			args: args{
				path:   `/update/counter/test/1`,
				method: http.MethodPost,
			},
			wantCode: 200,
		},
		{
			name: `Positive #2 (Metric gauge success)`,
			args: args{
				path:   `/update/gauge/test/1.2`,
				method: http.MethodPost,
			},
			wantCode: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			g := r.Group(`/`)
			s := repository.CreateRepository()

			RegisterAppHandler(g, s)

			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
			if tt.wantErrBody != "" {
				assert.Equal(t, tt.wantErrBody, strings.TrimSuffix(rec.Body.String(), "\n"))
				return
			}
		})
	}
}
