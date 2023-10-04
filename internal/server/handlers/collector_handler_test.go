package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCollectorHandler_ServeHTTP(t *testing.T) {
	type args struct {
		path   string
		method string
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantBody string
	}{
		{
			name: `Negative #1 (Without params)`,
			args: args{
				path:   `/update/`,
				method: http.MethodPost,
			},
			wantCode: 404,
			wantBody: `404 page not found`,
		},
		{
			name: `Negative #2 (Invalid type)`,
			args: args{
				path:   `/update/oops/name/1`,
				method: http.MethodPost,
			},
			wantCode: 400,
			wantBody: `type not found`,
		},
		{
			name: `Negative #3 (Counter fail)`,
			args: args{
				path:   `/update/counter/name/1.1`,
				method: http.MethodPost,
			},
			wantCode: 400,
			wantBody: `parse error`,
		},
		{
			name: `Positive #1 (Counter)`,
			args: args{
				path:   `/update/counter/name/1`,
				method: http.MethodPost,
			},
			wantCode: 200,
			wantBody: ``,
		},
		{
			name: `Positive #2 (Gauge)`,
			args: args{
				path:   `/update/gauge/name/1.1`,
				method: http.MethodPost,
			},
			wantCode: 200,
			wantBody: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			s := repository.CreateRepository()

			RegisterCollectorHandler(r, s)

			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSuffix(rec.Body.String(), "\n"))
		})
	}
}
