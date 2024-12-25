package rlutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHostLimiter(t *testing.T) {
	cases := []struct {
		name                string
		host                string
		reqLimitPerHost     map[string]int
		expectedReqLimit    int
		expectedToBeLimited bool
	}{
		{
			name:                "Host is limited",
			host:                "api.example.com",
			reqLimitPerHost:     map[string]int{},
			expectedReqLimit:    5,
			expectedToBeLimited: true,
		},
		{
			name: "when host is more limited",
			host: "api.example.com",
			reqLimitPerHost: map[string]int{
				"api.example.com": 1,
			},
			expectedReqLimit:    1,
			expectedToBeLimited: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockCounter := new(MockCounter)
			mockCounter.On("Get", mock.Anything, mock.Anything).Return(1, nil)
			mockCounter.On("Increment", mock.Anything, mock.Anything).Return(nil)

			reqLimit := 5
			windowLen := time.Minute

			limiter := NewHostLimiter(
				reqLimit,
				windowLen,
				nil,
				nil,
			)
			limiter.ReqLimitPerHost = tc.reqLimitPerHost

			// Using the mock counter instead of the real one.
			limiter.Counter = mockCounter

			req := httptest.NewRequest(http.MethodGet, "http://"+tc.host, nil)
			rule, err := limiter.Rule(req)
			assert.NoError(t, err)

			if tc.expectedToBeLimited {
				assert.NotNil(t, rule)
				assert.Equal(t, tc.host, rule.Key)
				assert.Equal(t, tc.expectedReqLimit, rule.ReqLimit)
				assert.Equal(t, windowLen, rule.WindowLen)
			} else {
				assert.Nil(t, rule)
			}
		})
	}
}
