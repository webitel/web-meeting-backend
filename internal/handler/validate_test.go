package handler

import (
	"testing"
)

func Test_validateURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		// Позитивні сценарії
		{
			name:    "valid https url",
			rawURL:  "https://google.com",
			wantErr: false,
		},
		{
			name:    "valid http url with path",
			rawURL:  "http://example.com/api/v1",
			wantErr: false,
		},
		{
			name:    "url with port",
			rawURL:  "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "url with query params",
			rawURL:  "https://search.com?q=golang",
			wantErr: false,
		},

		// Негативні сценарії
		{
			name:    "empty string",
			rawURL:  "",
			wantErr: true,
		},
		{
			name:    "just a string",
			rawURL:  "not-a-url",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			rawURL:  "google.com",
			wantErr: true,
		},
		{
			name:    "missing host",
			rawURL:  "https://",
			wantErr: true,
		},
		{
			name:    "relative path",
			rawURL:  "/path/to/resource",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			rawURL:  "http://go ogle.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
