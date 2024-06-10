package rlutils

import (
	"net/http"
	"strings"
	"time"

	"github.com/2manymws/rl"
)

type RequestPathLimiter struct {
	requestPathContains []string
	requestPathPrefixes []string
	requestPathSuffixes []string
	ignorePathContains  []string
	ignorePathPrefixes  []string
	ignorePathSuffixes  []string
	key                 string
	BaseLimiter
}

// リクエストパスごとにリクエスト数を制限する
// 制限単位はホスト名 + リクエストパス
func NewRequestPathLimiter(
	requestPathContains []string,
	requestPathPrefixes []string,
	requestPathSuffixes []string,
	reqLimit int,
	windowLen time.Duration,
	key string,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
	setter ...Option,

) (*RequestPathLimiter, error) {
	err := validateKey(key)
	if err != nil {
		return nil, err
	}

	return &RequestPathLimiter{
		requestPathContains: requestPathContains,
		requestPathPrefixes: requestPathPrefixes,
		requestPathSuffixes: requestPathSuffixes,
		key:                 key,
		BaseLimiter: NewBaseLimiter(
			reqLimit,
			windowLen,
			onRequestLimit,
			setter...,
		),
	}, nil
}

func (l *RequestPathLimiter) Name() string {
	return "request_path_limiter"
}

func (l *RequestPathLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.IsTargetRequest(r) {
		return &rl.Rule{ReqLimit: -1}, nil
	}
	for _, st := range []struct {
		path []string
		f    func(string, string) bool
	}{
		{l.requestPathPrefixes, strings.HasPrefix},
		{l.requestPathSuffixes, strings.HasSuffix},
		{l.requestPathContains, strings.Contains},
	} {
		if len(st.path) > 0 {
			for _, path := range st.path {
				if st.f(r.URL.Path, path) {
					return &rl.Rule{
						Key:       fillKey(r, l.key) + path,
						ReqLimit:  l.reqLimit,
						WindowLen: l.windowLen,
					}, nil
				}
			}
		}
	}
	return &rl.Rule{ReqLimit: -1}, nil
}

func (l *RequestPathLimiter) OnRequestLimit(r *rl.Context) http.HandlerFunc {
	return l.onRequestLimit(r, l.Name())
}
