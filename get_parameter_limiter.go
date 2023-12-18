package rlutils

import (
	"net/http"
	"time"

	"github.com/2manymws/rl"
)

type GetParameterLimiter struct {
	getParameters map[string]string
	BaseLimiter
}

// Getパラメーターごとにリクエスト数を制限する
// 制限単位はホスト名とGetパラメーター
func NewGetParameterLimiter(
	getParameters map[string]string,
	reqLimit int,
	windowLen time.Duration,
	targetExtensions []string,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
) *GetParameterLimiter {
	return &GetParameterLimiter{
		getParameters: getParameters,
		BaseLimiter: NewBaseLimiter(
			reqLimit,
			windowLen,
			targetExtensions,
			onRequestLimit,
		),
	}
}

func (l *GetParameterLimiter) Name() string {
	return "get_parameter_limiter"
}

func (l *GetParameterLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.IsTargetRequest(r) {
		return &rl.Rule{ReqLimit: -1}, nil
	}
	for k, v := range l.getParameters {
		if r.URL.Query().Get(k) == v {
			return &rl.Rule{
				Key:       r.Host + "/" + k + "=" + v,
				ReqLimit:  l.reqLimit,
				WindowLen: l.windowLen,
			}, nil
		}
	}

	return &rl.Rule{ReqLimit: -1}, nil
}

func (l *GetParameterLimiter) OnRequestLimit(r *rl.Context) http.HandlerFunc {
	return l.onRequestLimit(r, l.Name())
}
