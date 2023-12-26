package rlutils

import (
	"net/http/httptest"
	"testing"
)

func TestBaseLimiter_isTargetExtensions(t *testing.T) {
	tests := []struct {
		name             string
		targetExtensions []string
		requestPath      string
		want             bool
	}{
		{
			name:             "Target extension .txt",
			targetExtensions: []string{".txt"},
			requestPath:      "/file.txt",
			want:             true,
		},
		{
			name:             "Target extension .txt",
			targetExtensions: []string{"txt"},
			requestPath:      "/file.txt",
			want:             true,
		},
		{
			name:             "Target extension .jpg among others",
			targetExtensions: []string{".png", ".jpg", ".gif"},
			requestPath:      "/image.jpg",
			want:             true,
		},
		{
			name:             "Target extension case insensitive",
			targetExtensions: []string{".TXT"},
			requestPath:      "/file.txt",
			want:             true,
		},
		{
			name:             "No target extensions set",
			targetExtensions: []string{},
			requestPath:      "/file.txt",
			want:             true,
		},
		{
			name:             "Extension not in target list",
			targetExtensions: []string{".png", ".jpg", ".gif"},
			requestPath:      "/file.txt",
			want:             false,
		},
		{
			name:             "Request path without extension",
			targetExtensions: []string{".txt"},
			requestPath:      "/file",
			want:             false,
		},
		{
			name:             "null extention",
			targetExtensions: []string{""},
			requestPath:      "/file",
			want:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a BaseLimiter with specific target extensions.
			bl := NewBaseLimiter(
				0,
				0,
				tt.targetExtensions,
				nil,
			)
			// Create an HTTP request with the test case path.
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			// Check if the request extension is in the target list.
			got := bl.isTargetExtensions(req)
			if got != tt.want {
				t.Errorf("isTargetExtensions() = %v, want %v", got, tt.want)
			}
		})
	}
}
