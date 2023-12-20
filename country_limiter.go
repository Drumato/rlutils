package rlutils

// limit from ip with maxMindDB

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/2manymws/rl"
	maxminddb "github.com/oschwald/maxminddb-golang"
)

type CountryLimiter struct {
	db            *maxminddb.Reader
	countries     map[string]struct{}
	skipCountries map[string]struct{}
	BaseLimiter
}

// 国別のリクエスト数を制限する
// 制限単位はIPアドレス
func NewCountryLimiter(
	dbPath string,
	countries []string,
	skipCountries []string,
	reqLimit int,
	windowLen time.Duration,
	targetExtensions []string,
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
) (*CountryLimiter, error) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}
	cm := map[string]bool{}
	scm := map[string]bool{}

	for _, c := range countries {
		cm[c] = true
	}

	for _, c := range skipCountries {
		if c == "*" {
			return nil, fmt.Errorf("invalid skip country: %s", c)
		}
		scm[c] = true
	}
	return &CountryLimiter{
		db:            db,
		countries:     cm,
		skipCountries: scm,
		BaseLimiter: NewBaseLimiter(
			reqLimit,
			windowLen,
			targetExtensions,
			onRequestLimit,
		),
	}, nil
}

func (l *CountryLimiter) Name() string {
	return "country_limiter"
}

func (l *CountryLimiter) Rule(r *http.Request) (*rl.Rule, error) {
	if !l.IsTargetRequest(r) {
		return &rl.Rule{ReqLimit: -1}, nil
	}

	remoteAddr := strings.Split(r.RemoteAddr, ":")[0]
	country, err := l.country(remoteAddr)
	if err != nil {
		return nil, err
	}

	limit := &rl.Rule{
		Key:       remoteAddr,
		ReqLimit:  l.reqLimit,
		WindowLen: l.windowLen,
	}
	noLimit := &rl.Rule{ReqLimit: -1}

	if country == "" {
		return noLimit, nil
	}

	if _, ok := l.skipCountries[country]; ok {
		return noLimit, nil
	}

	if _, ok := l.countries["*"]; ok {
		return limit, nil

	}

	if _, ok := l.countries[country]; ok {
		return &rl.Rule{
			Key:       remoteAddr,
			ReqLimit:  l.reqLimit,
			WindowLen: l.windowLen,
		}, nil
	}
	return noLimit, nil
}

func (l *CountryLimiter) country(remoteAddr string) (string, error) {
	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	err := l.db.Lookup(net.ParseIP(remoteAddr), &record)
	if err != nil {
		return "", err
	}

	return record.Country.ISOCode, nil
}

func (l *CountryLimiter) OnRequestLimit(r *rl.Context) http.HandlerFunc {
	return l.onRequestLimit(r, l.Name())
}
