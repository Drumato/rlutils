package rlutils

import (
	"net/http"
	"strings"
	"time"

	"github.com/2manymws/rl"
)

type IPLimiter struct {
	BaseLimiter
}

// ホストごとにリクエスト数を制限する
// 制限単位はIP
func NewIPLimiter(
	reqLimit int,
	windowLen time.Duration,
	targetExtensions []string,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
) *IPLimiter {
	return &IPLimiter{
		BaseLimiter: NewBaseLimiter(
			reqLimit,
			windowLen,
			targetExtensions,
			onRequestLimit,
		),
	}
}

func (l *IPLimiter) Name() string {
	return "ip_limiter"
}

func (l *IPLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.IsTargetRequest(r) {
		return &rl.Rule{ReqLimit: -1}, nil
	}
	remoteAddr := strings.Split(r.RemoteAddr, ":")[0]
	return &rl.Rule{
		Key:       remoteAddr,
		ReqLimit:  l.reqLimit,
		WindowLen: l.windowLen,
	}, nil
}

func (l *IPLimiter) OnRequestLimit(r *rl.Context) http.HandlerFunc {
	return l.onRequestLimit(r, l.Name())
}
