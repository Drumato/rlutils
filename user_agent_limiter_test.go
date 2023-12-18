package rlutils

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCounter struct {
	mock.Mock
}

func (m *MockCounter) Get(key string, window time.Time) (int, error) {
	args := m.Called(key, window)
	return args.Int(0), args.Error(1)
}

func (m *MockCounter) Increment(key string, window time.Time) error {
	args := m.Called(key, window)
	return args.Error(0)
}

func TestUserAgentLimiter(t *testing.T) {
	mockCounter := new(MockCounter)
	mockCounter.On("Get", mock.Anything, mock.Anything).Return(1, nil)
	mockCounter.On("Increment", mock.Anything, mock.Anything).Return(nil)

	userAgents := []string{"TestBot", "SuperBot"}
	reqLimit := 5
	windowLen := 1 * time.Minute

	limiter := NewUserAgentLimiter(
		userAgents,
		reqLimit,
		windowLen,
		nil,
		nil,
	)
	limiter.Counter = mockCounter

	// Create a request with matching User-Agent
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("User-Agent", "TestBot 1.0")

	// Check if the correct rule is returned for a matching User-Agent string
	rule, err := limiter.Rule(req)
	assert.NoError(t, err)
	assert.NotNil(t, rule)
	if rule != nil {
		assert.Equal(t, reqLimit, rule.ReqLimit)
		assert.Equal(t, windowLen, rule.WindowLen)
		assert.Equal(t, "TestBot", rule.Key)
	}

	// Create a request with a non-matching User-Agent
	nonMatchingReq := httptest.NewRequest("GET", "http://example.com", nil)
	nonMatchingReq.Header.Set("User-Agent", "UnknownBot 1.0")

	// Check if no rule is returned for a non-matching User-Agent string
	noRule, err := limiter.Rule(nonMatchingReq)
	assert.NoError(t, err)
	assert.Equal(t, noRule.ReqLimit, -1)
}
