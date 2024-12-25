package rlutils

import (
	"net/http"
	"time"

	"github.com/2manymws/rl"
)

type HostLimiter struct {
	BaseLimiter
	// ReqLimitPerHost はホストごとに特別なしきい値を設定できるようにします
	ReqLimitPerHost map[string]int
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
		ReqLimitPerHost: make(map[string]int),
	}
}

func (l *HostLimiter) Name() string {
	return "host_limiter"
}

func (l *HostLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.IsTargetRequest(r) {
		return &rl.Rule{ReqLimit: -1}, nil
	}

	reqLimit := l.reqLimit
	if v, ok := l.ReqLimitPerHost[r.Host]; ok {
		reqLimit = v
	}
	return &rl.Rule{
		Key:       r.Host,
		ReqLimit:  reqLimit,
		WindowLen: l.windowLen,
	}, nil
}

func (l *HostLimiter) OnRequestLimit(r *rl.Context) http.HandlerFunc {
	return l.onRequestLimit(r, l.Name())
}
