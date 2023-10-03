package collector

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Jourloy/go-metrics-collector/cmd/server/storage/repository"
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
			name: `Negative #1 (Without prefix)`,
			args: args{
				path:   `/update`,
				method: http.MethodPost,
			},
			wantCode: 404,
			wantBody: `not found prefix`,
		},
		{
			name: `Negative #2 (Without params)`,
			args: args{
				path:   `/update/`,
				method: http.MethodPost,
			},
			wantCode: 404,
			wantBody: `length of url params is not 3`,
		},
		{
			name: `Negative #3 (Invalid type)`,
			args: args{
				path:   `/update/oops/name/1`,
				method: http.MethodPost,
			},
			wantCode: 400,
			wantBody: `type not found`,
		},
		{
			name: `Negative #4 (Counter fail)`,
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
			s := repository.CreateRepository()
			c := &CollectorHandler{storage: s}

			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			rec := httptest.NewRecorder()

			c.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSuffix(rec.Body.String(), "\n"))
		})
	}
}
