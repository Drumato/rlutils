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
		{name: "Target extension .txt",
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
				nil,
				TargetExtensions(tt.targetExtensions),
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

func TestBaseLimiter_isTargetMethod(t *testing.T) {
	tests := []struct {
		name          string
		targetMethods []string
		requestMethod string
		want          bool
	}{
		{
			name:          "Target method GET",
			targetMethods: []string{"GET"},
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "Target method POST",
			targetMethods: []string{"POST"},
			requestMethod: "POST",
			want:          true,
		},
		{
			name:          "Method not in target list",
			targetMethods: []string{"GET", "POST"},
			requestMethod: "DELETE",
			want:          false,
		},
		{
			name:          "Case insensitive method check",
			targetMethods: []string{"get"},
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "No target methods set",
			targetMethods: []string{},
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "Multiple methods, target hit",
			targetMethods: []string{"GET", "POST", "PUT"},
			requestMethod: "PUT",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a BaseLimiter with specific target methods.
			bl := NewBaseLimiter(
				0,
				0,
				nil,
				TargetMethods(tt.targetMethods),
			)
			req := httptest.NewRequest(tt.requestMethod, "/some/path", nil)
			got := bl.isTargetMethod(req)
			if got != tt.want {
				t.Errorf("isTargetMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}
