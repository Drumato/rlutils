package rlutils

import (
	"net/http"
	"time"

	"github.com/2manymws/rl"
)

type HostLimiter struct {
	BaseLimiter
}

// ホストごとにリクエスト数を制限する
// 制限単位はホスト名
func NewHostLimiter(
	reqLimit int,
	windowLen time.Duration,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
	setter ...Option,
) *HostLimiter {
	return &HostLimiter{
		BaseLimiter: NewBaseLimiter(
			reqLimit,
			windowLen,
			onRequestLimit,
			setter...,
		),
	}
}

func (l *HostLimiter) Name() string {
	return "host_limiter"
}

func (l *HostLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.IsTargetRequest(r) {
		return &rl.Rule{ReqLimit: -1}, nil
	}
	return &rl.Rule{
		Key:       r.Host,
		ReqLimit:  l.reqLimit,
		WindowLen: l.windowLen,
	}, nil
}

func (l *HostLimiter) OnRequestLimit(r *rl.Context) http.HandlerFunc {
	return l.onRequestLimit(r, l.Name())
}
