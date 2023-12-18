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
		expectedToBeLimited bool
		expectedKey         string
	}{
		{
			name: "Request with matching get parameter should be limited",
			getParameters: map[string]string{
				"token": "123456",
			},
			queryString:         "?token=123456",
			expectedToBeLimited: true,
			expectedKey:         "example.com/token=123456",
		},
		{
			name: "Request without matching get parameter should not be limited",
			getParameters: map[string]string{
				"token": "123456",
			},
			queryString:         "?token=abcdef",
			expectedToBeLimited: false,
		},
		{
			name: "Request with multiple get parameters, one matches should be limited",
			getParameters: map[string]string{
				"token":     "123456",
				"sessionId": "ABCDEF",
			},
			queryString:         "?token=123456&sessionId=XYZ",
			expectedToBeLimited: true,
			expectedKey:         "example.com/token=123456",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockCounter := new(MockCounter)
			mockCounter.On("Get", mock.Anything, mock.Anything).Return(1, nil)
			mockCounter.On("Increment", mock.Anything, mock.Anything).Return(nil)

			reqLimit := 5
			windowLen := time.Minute

			limiter := NewGetParameterLimiter(
				tc.getParameters,
				reqLimit,
				windowLen,
				nil,
				nil,
			)

			// Using the mock counter instead of the real one.
			limiter.Counter = mockCounter

			req := httptest.NewRequest(http.MethodGet, "http://example.com"+tc.queryString, nil)
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
