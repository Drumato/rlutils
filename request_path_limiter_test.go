package rlutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequestPathLimiter(t *testing.T) {
	cases := []struct {
		name                string
		contains            []string
		prefixes            []string
		suffixes            []string
		ignoreContains      []string
		ignorePrefixes      []string
		ignoreSuffixes      []string
		path                string
		expectedToBeLimited bool
		expectedKey         string
	}{
		{
			name:                "Path contains limited segment",
			contains:            []string{"/user"},
			prefixes:            []string{"/api/"},
			suffixes:            []string{"/details"},
			path:                "/accounts/user/profile",
			expectedToBeLimited: true,
			expectedKey:         "example.com/user",
		},
		{
			name:                "Path starts with limited prefix",
			contains:            []string{"user"},
			prefixes:            []string{"/api/"},
			suffixes:            []string{"/details"},
			path:                "/api/users/1",
			expectedToBeLimited: true,
			expectedKey:         "example.com/api/",
		},
		{
			name:                "Path ends with limited suffix",
			contains:            []string{"user"},
			prefixes:            []string{"/api/"},
			suffixes:            []string{"/details"},
			path:                "/users/1/details",
			expectedToBeLimited: true,
			expectedKey:         "example.com/details",
		},
		{
			name:                "Path does not match any criteria",
			contains:            []string{"user"},
			prefixes:            []string{"/api/"},
			suffixes:            []string{"/details"},
			path:                "/about",
			expectedToBeLimited: false,
		},
		{
			name:                "Ignore path contains limited segment",
			contains:            []string{"/abc"},
			prefixes:            []string{"/abcd"},
			suffixes:            []string{"/abcde"},
			path:                "/abcdefg",
			ignoreContains:      []string{"ab"},
			expectedToBeLimited: false,
			expectedKey:         "",
		},
		{
			name:                "Ignore path starts with limited prefix",
			contains:            []string{"/abc"},
			prefixes:            []string{"/abcd"},
			suffixes:            []string{"/abcde"},
			path:                "/abcdefg",
			ignorePrefixes:      []string{"/a"},
			expectedToBeLimited: false,
			expectedKey:         "",
		},
		{
			name:                "Ignore path ends with limited suffix",
			contains:            []string{"/abc"},
			prefixes:            []string{"/abcd"},
			suffixes:            []string{"/abcde"},
			path:                "/abcdefg",
			ignoreSuffixes:      []string{"g"},
			expectedToBeLimited: false,
			expectedKey:         "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockCounter := new(MockCounter)
			mockCounter.On("Get", mock.Anything, mock.Anything).Return(1, nil)
			mockCounter.On("Increment", mock.Anything, mock.Anything).Return(nil)

			reqLimit := 5
			windowLen := time.Minute

			limiter := NewRequestPathLimiter(
				tc.contains,
				tc.prefixes,
				tc.suffixes,
				tc.ignoreContains,
				tc.ignorePrefixes,
				tc.ignoreSuffixes,
				reqLimit,
				windowLen,
				nil,
				nil,
			)

			// Using the mock counter instead of the real one.
			limiter.Counter = mockCounter

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rule, err := limiter.Rule(req)
			assert.NoError(t, err)

			if tc.expectedToBeLimited {
				assert.NotNil(t, rule)
				assert.Equal(t, tc.expectedKey, rule.Key)
				assert.Equal(t, reqLimit, rule.ReqLimit)
				assert.Equal(t, windowLen, rule.WindowLen)
			} else {
				assert.Equal(t, -1, rule.ReqLimit)
			}
		})
	}
}
