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
	targetExtensions []string,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
) *HostLimiter {
	return &HostLimiter{
		BaseLimiter: NewBaseLimiter(
			reqLimit,
			windowLen,
			targetExtensions,
			onRequestLimit,
		),
	}
}

func (l *HostLimiter) Name() string {
	return "host_limiter"
}

func (l *HostLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.isTargetRequest(r) {
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
