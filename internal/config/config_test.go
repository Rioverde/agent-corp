package config

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestSetInt32FromVault(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		key     string
		want    int32
		wantErr bool
	}{
		{
			name: "string number",
			data: map[string]any{"K": "42"},
			key:  "K",
			want: 42,
		},
		{
			name: "json.Number",
			data: map[string]any{"K": json.Number("100")},
			key:  "K",
			want: 100,
		},
		{
			name: "float64",
			data: map[string]any{"K": float64(7)},
			key:  "K",
			want: 7,
		},
		{
			name: "int",
			data: map[string]any{"K": 11},
			key:  "K",
			want: 11,
		},
		{
			name: "missing key keeps zero",
			data: map[string]any{},
			key:  "K",
			want: 0,
		},
		{
			name:    "invalid string",
			data:    map[string]any{"K": "abc"},
			key:     "K",
			wantErr: true,
		},
		{
			name:    "overflow int32",
			data:    map[string]any{"K": int64(math.MaxInt32) + 1},
			key:     "K",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			data:    map[string]any{"K": []byte("nope")},
			key:     "K",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int32
			err := setInt32FromVault(tt.data, tt.key, &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("setInt32FromVault err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}

func TestSetDurationFromVault(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		key     string
		want    time.Duration
		wantErr bool
	}{
		{
			name: "valid duration",
			data: map[string]any{"K": "5m"},
			key:  "K",
			want: 5 * time.Minute,
		},
		{
			name: "missing key keeps zero",
			data: map[string]any{},
			key:  "K",
			want: 0,
		},
		{
			name:    "invalid string",
			data:    map[string]any{"K": "not-a-duration"},
			key:     "K",
			wantErr: true,
		},
		{
			name:    "non-string",
			data:    map[string]any{"K": 123},
			key:     "K",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got time.Duration
			err := setDurationFromVault(tt.data, tt.key, &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("setDurationFromVault err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
