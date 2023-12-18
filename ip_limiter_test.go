package rlutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIPLimiter(t *testing.T) {
	cases := []struct {
		name                string
		remoteAddr          string
		expectedToBeLimited bool
		expectedKey         string
	}{
		{
			name:                "IP is limited",
			remoteAddr:          "10.0.0.1",
			expectedToBeLimited: true,
			expectedKey:         "10.0.0.1",
		},
		{
			name:                "IP With Port is limited",
			remoteAddr:          "10.0.0.1:12345",
			expectedToBeLimited: true,
			expectedKey:         "10.0.0.1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockCounter := new(MockCounter)
			mockCounter.On("Get", mock.Anything, mock.Anything).Return(1, nil)
			mockCounter.On("Increment", mock.Anything, mock.Anything).Return(nil)

			reqLimit := 5
			windowLen := time.Minute

			limiter := NewIPLimiter(
				reqLimit,
				windowLen,
				nil,
				nil,
			)

			// Using the mock counter instead of the real one.
			limiter.Counter = mockCounter

			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			req.RemoteAddr = tc.remoteAddr
			rule, err := limiter.Rule(req)
			assert.NoError(t, err)

			if tc.expectedToBeLimited {
				assert.NotNil(t, rule)
				assert.Equal(t, tc.expectedKey, rule.Key)
				assert.Equal(t, reqLimit, rule.ReqLimit)
				assert.Equal(t, windowLen, rule.WindowLen)
			} else {
				assert.Nil(t, rule)
			}
		})
	}
}
