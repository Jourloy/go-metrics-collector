package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_live(t *testing.T) {
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
			name: `Negative #1`,
			args: args{
				path:   `/live`,
				method: http.MethodPost,
			},
			wantCode: 405,
			wantBody: `method not allowed`,
		},
		{
			name: `Positive #1`,
			args: args{
				path:   `/live`,
				method: http.MethodGet,
			},
			wantCode: 200,
			wantBody: `Live`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			rec := httptest.NewRecorder()

			live(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSuffix(rec.Body.String(), "\n"))
		})
	}
}
