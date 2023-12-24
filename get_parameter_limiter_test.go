package rlutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCounter implementation as previously described.

func TestGetParameterLimiter(t *testing.T) {
	cases := []struct {
		name                string
		getParameters       map[string]string
		queryString         string
		key                 string
		expectedToBeLimited bool
		expectedKey         string
	}{
		{
			name: "Request with matching get parameter should be limited",
			getParameters: map[string]string{
				"token": "123456",
			},
			queryString:         "?token=123456",
			key:                 "host",
			expectedToBeLimited: true,
			expectedKey:         "example.com/token=123456",
		},
		{
			name: "Request without matching get parameter should not be limited",
			getParameters: map[string]string{
				"token": "123456",
			},
			queryString:         "?token=abcdef",
			key:                 "host",
			expectedToBeLimited: false,
		},
		{
			name: "Request with multiple get parameters, one matches should be limited",
			getParameters: map[string]string{
				"token":     "123456",
				"sessionId": "ABCDEF",
			},
			queryString:         "?token=123456&sessionId=XYZ",
			key:                 "host",
			expectedToBeLimited: true,
			expectedKey:         "example.com/token=123456",
		},
		{
			name: "Request with matching get parameter should be limited with remote ip",
			getParameters: map[string]string{
				"token": "123456",
			},
			queryString:         "?token=123456",
			key:                 "remote_addr",
			expectedToBeLimited: true,
			expectedKey:         "127.0.0.1/token=123456",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockCounter := new(MockCounter)
			mockCounter.On("Get", mock.Anything, mock.Anything).Return(1, nil)
			mockCounter.On("Increment", mock.Anything, mock.Anything).Return(nil)

			reqLimit := 5
			windowLen := time.Minute

			limiter, _ := NewGetParameterLimiter(
				tc.getParameters,
				reqLimit,
				windowLen,
				tc.key,
				nil,
				nil,
			)

			// Using the mock counter instead of the real one.
			limiter.Counter = mockCounter

			req := httptest.NewRequest(http.MethodGet, "http://example.com"+tc.queryString, nil)
			req.RemoteAddr = "127.0.0.1:12345"
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
