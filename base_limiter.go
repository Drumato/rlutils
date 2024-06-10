package rlutils

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/2manymws/rl"
	"github.com/2manymws/rl/counter"
)

const (
	RemoteAddrKey = "remote_addr"
	HostKey       = "host"
)

type Options struct {
	TargetExtensions   map[string]struct{}
	TargetMethods      map[string]struct{}
	IgnorePathContains []string
	IgnorePathPrefixes []string
	IgnorePathSuffixes []string
}

type Option func(*Options)

type BaseLimiter struct {
	reqLimit           int `mapstructure:"req_limit"`
	windowLen          time.Duration
	targetExtensions   map[string]struct{}
	targetMethods      map[string]struct{}
	ignorePathContains []string
	ignorePathPrefixes []string
	ignorePathSuffixes []string
	onRequestLimit     func(*rl.Context, string) http.HandlerFunc
	rl.Counter
}

func NewBaseLimiter(
	reqLimit int,
	windowLen time.Duration,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
	setters ...Option,
) BaseLimiter {
	ttl := windowLen * 2 // 最低2回分のウィンドウ分のカウンタを維持する

	options := Options{}

	for _, setter := range setters {
		if setter != nil {
			setter(&options)
		}
	}

	return BaseLimiter{
		reqLimit:           reqLimit,
		windowLen:          windowLen,
		Counter:            counter.New(ttl),
		targetExtensions:   options.TargetExtensions,
		targetMethods:      options.TargetMethods,
		onRequestLimit:     onRequestLimit,
		ignorePathContains: options.IgnorePathContains,
		ignorePathPrefixes: options.IgnorePathPrefixes,
		ignorePathSuffixes: options.IgnorePathSuffixes,
	}
}

func TargetExtensions(targetExtensions []string) Option {
	return func(args *Options) {
		args.TargetExtensions = make(map[string]struct{}, len(targetExtensions))
		if len(targetExtensions) > 0 {
			for _, ext := range targetExtensions {
				if len(ext) > 0 && ext[0] != '.' {
					ext = "." + ext
				}
				args.TargetExtensions[strings.ToLower(ext)] = struct{}{}
			}
		}
	}
}

func TargetMethods(targetMethods []string) Option {
	return func(args *Options) {
		args.TargetMethods = make(map[string]struct{}, len(targetMethods))
		if len(targetMethods) > 0 {
			for _, method := range targetMethods {
				args.TargetMethods[strings.ToLower(method)] = struct{}{}
			}
		}
	}
}

func IgnorePathContains(ignorePathContains []string) Option {
	return func(args *Options) {
		args.IgnorePathContains = ignorePathContains
	}
}

func IgnorePathPrefixes(ignorePathPrefixes []string) Option {
	return func(args *Options) {
		args.IgnorePathPrefixes = ignorePathPrefixes
	}
}

func IgnorePathSuffixes(ignorePathSuffixes []string) Option {
	return func(args *Options) {
		args.IgnorePathSuffixes = ignorePathSuffixes
	}
}

func (l *BaseLimiter) ShouldSetXRateLimitHeaders(r *rl.Context) bool {
	return false
}

func (l *BaseLimiter) Name() string {
	return "base_limiter"
}

func (l *BaseLimiter) IsTargetRequest(r *http.Request) bool {
	return l.isTargetExtensions(r) && l.isTargetMethod(r) && l.isTargetPath(r)
}

func (l *BaseLimiter) isTargetExtensions(r *http.Request) bool {
	if len(l.targetExtensions) == 0 {
		return true
	}
	extension := strings.ToLower(filepath.Ext(r.URL.Path))
	_, ok := l.targetExtensions[extension]
	return ok
}

func (l *BaseLimiter) isTargetMethod(r *http.Request) bool {
	if len(l.targetMethods) == 0 {
		return true
	}
	_, ok := l.targetMethods[strings.ToLower(r.Method)]
	return ok
}

func (l *BaseLimiter) isTargetPath(r *http.Request) bool {
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
					return false
				}
			}
		}
	}

	return true
}
func validateKey(key string) error {

	for _, k := range []string{RemoteAddrKey, HostKey} {
		if k == key {
			return nil
		}
	}
	return fmt.Errorf("invalid key: %s", key)
}

func fillKey(r *http.Request, key string) string {
	if key == RemoteAddrKey {
		remoteAddr := strings.Split(r.RemoteAddr, ":")[0]
		return remoteAddr
	}
	return r.Host
}
