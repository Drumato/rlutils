package rlutils

import (
	"net/http"
	"strings"
	"time"

	"github.com/2manymws/rl"
)

type UserAgentLimiter struct {
	userAgents []string
	BaseLimiter
}

// ユーザーエージェントごとにリクエスト数を制限する
// 制限単位はユーザーエージェント
func NewUserAgentLimiter(
	userAgents []string,
	reqLimit int,
	windowLen time.Duration,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
	setter ...Option,
) *UserAgentLimiter {
	return &UserAgentLimiter{
		userAgents: userAgents,
		BaseLimiter: NewBaseLimiter(
			reqLimit,
			windowLen,
			onRequestLimit,
			setter...,
		),
	}
}

func (l *UserAgentLimiter) Name() string {
	return "user_agent_limiter"
}

func (l *UserAgentLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.IsTargetRequest(r) {
		return &rl.Rule{ReqLimit: -1}, nil
	}
	for _, ua := range l.userAgents {
		if strings.Contains(r.UserAgent(), ua) {
			return &rl.Rule{
				Key:       ua,
				ReqLimit:  l.reqLimit,
				WindowLen: l.windowLen,
			}, nil
		}
	}
	return &rl.Rule{ReqLimit: -1}, nil
}

func (l *UserAgentLimiter) OnRequestLimit(r *rl.Context) http.HandlerFunc {
	return l.onRequestLimit(r, l.Name())
}
