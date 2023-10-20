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

func TestRegisterValueHandler(t *testing.T) {
	type args struct {
		path string
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
				path: `/value/`,
			},
			wantCode: 404,
			wantBody: `404 page not found`,
		},
		{
			name: `Negative #2 (Unknown params)`,
			args: args{
				path: `/value/H`,
			},
			wantCode: 404,
			wantBody: `404 page not found`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			g := r.Group(`/value`)
			s := repository.CreateRepository()

			RegisterValueHandler(g, s)

			req := httptest.NewRequest(http.MethodGet, tt.args.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSuffix(rec.Body.String(), "\n"))
		})
	}
}
