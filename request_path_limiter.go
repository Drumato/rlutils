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
	BaseLimiter
}

// リクエストパスごとにリクエスト数を制限する
// 制限単位はホスト名 + リクエストパス
func NewRequestPathLimiter(
	requestPathContains []string,
	requestPathPrefixes []string,
	requestPathSuffixes []string,
	ignorePathContains []string,
	ignorePathPrefixes []string,
	ignorePathSuffixes []string,
	reqLimit int,
	windowLen time.Duration,
	targetExtensions []string,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
) *RequestPathLimiter {
	return &RequestPathLimiter{
		requestPathContains: requestPathContains,
		requestPathPrefixes: requestPathPrefixes,
		requestPathSuffixes: requestPathSuffixes,
		ignorePathContains:  ignorePathContains,
		ignorePathPrefixes:  ignorePathPrefixes,
		ignorePathSuffixes:  ignorePathSuffixes,
		BaseLimiter: NewBaseLimiter(
			reqLimit,
			windowLen,
			targetExtensions,
			onRequestLimit,
		),
	}
}

func (l *RequestPathLimiter) Name() string {
	return "request_path_limiter"
}

func (l *RequestPathLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.isTargetRequest(r) {
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
					ignored := false

					// 除外条件にマッチする場合は無視する
					for _, ignore := range []struct {
						path []string
						f    func(string, string) bool
					}{
						{l.ignorePathPrefixes, strings.HasPrefix},
						{l.ignorePathSuffixes, strings.HasSuffix},
						{l.ignorePathContains, strings.Contains},
					} {
						if len(ignore.path) > 0 {
							for _, ipath := range ignore.path {
								if ignore.f(r.URL.Path, ipath) {
									ignored = true
									break
								}
							}
							if ignored {
								break
							}
						}
					}
					if !ignored {
						return &rl.Rule{
							Key:       r.Host + path,
							ReqLimit:  l.reqLimit,
							WindowLen: l.windowLen,
						}, nil
					}

				}
			}
		}
	}
	return &rl.Rule{ReqLimit: -1}, nil
}

func (l *RequestPathLimiter) OnRequestLimit(r *rl.Context) http.HandlerFunc {
	return l.onRequestLimit(r, l.Name())
}
