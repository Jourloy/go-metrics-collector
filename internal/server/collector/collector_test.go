package collector

import (
	"testing"

	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestCollectorHandler_parseURL(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	type args struct {
		path string
	}
	type ParsedURL struct {
		Type  string
		Name  string
		Value string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		want           *ParsedURL
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: `Negative #1 (Without params)`,
			args: args{
				path: `/update/`,
			},
			want:           nil,
			wantErr:        true,
			wantErrMessage: `404 page not found`,
		},
		{
			name: `Negative #2 (Not enough params)`,
			args: args{
				path: `/update/counter/check`,
			},
			want:           nil,
			wantErr:        true,
			wantErrMessage: `404 page not found`,
		},
		{
			name: `Positive #1`,
			args: args{
				path: `/update/counter/check/1`,
			},
			want: &ParsedURL{
				Type:  `counter`,
				Name:  `check`,
				Value: `1`,
			},
			wantErr:        false,
			wantErrMessage: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CollectorHandler{
				storage: tt.fields.storage,
			}
			got, err := c.parseURL(tt.args.path)

			if err != nil {
				assert.True(t, tt.wantErr)
				assert.Equal(t, tt.wantErrMessage, err.Error())
			}

			if err == nil {
				assert.False(t, tt.wantErr)
				assert.Equal(t, tt.want.Type, got.Type)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Value, got.Value)
			}
		})
	}
}

func TestCollectorHandler_parseCounter(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	type args struct {
		param string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		want           int64
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: `Negative #1 (Without params)`,
			args: args{
				param: ``,
			},
			want:           0,
			wantErr:        true,
			wantErrMessage: `parse error`,
		},
		{
			name: `Negative #2 (Not int)`,
			args: args{
				param: `1.23`,
			},
			want:           0,
			wantErr:        true,
			wantErrMessage: `parse error`,
		},
		{
			name: `Positive #1`,
			args: args{
				param: `11`,
			},
			want:           11,
			wantErr:        false,
			wantErrMessage: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CollectorHandler{
				storage: tt.fields.storage,
			}
			got, err := c.parseCounter(tt.args.param)

			if err != nil {
				assert.True(t, tt.wantErr)
				assert.Equal(t, tt.wantErrMessage, err.Error())
			}

			if err == nil {
				assert.False(t, tt.wantErr)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCollectorHandler_parseGauge(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	type args struct {
		param string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		want           float64
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: `Negative #1 (Without params)`,
			args: args{
				param: ``,
			},
			want:           0,
			wantErr:        true,
			wantErrMessage: `parse error`,
		},
		{
			name: `Negative #2 (Invalid params)`,
			args: args{
				param: `12.,,12`,
			},
			want:           0,
			wantErr:        true,
			wantErrMessage: `parse error`,
		},
		{
			name: `Positive #1`,
			args: args{
				param: `11`,
			},
			want:           11,
			wantErr:        false,
			wantErrMessage: ``,
		},
		{
			name: `Positive #2`,
			args: args{
				param: `1.1`,
			},
			want:           1.1,
			wantErr:        false,
			wantErrMessage: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CollectorHandler{
				storage: tt.fields.storage,
			}
			got, err := c.parseGauge(tt.args.param)

			if err != nil {
				assert.True(t, tt.wantErr)
				assert.Equal(t, tt.wantErrMessage, err.Error())
			}

			if err == nil {
				assert.False(t, tt.wantErr)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
