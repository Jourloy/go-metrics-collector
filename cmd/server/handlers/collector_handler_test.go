package handlers

import (
	"net/http"
	"testing"
)

func TestRegisterCollectorHandler(t *testing.T) {
	type args struct {
		mux *http.ServeMux
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: `Positive #1`,
			args: args{
				mux: http.NewServeMux(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterCollectorHandler(tt.args.mux)
		})
	}
}
