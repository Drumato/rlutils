package rlutils

// limit from ip with maxMindDB

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/2manymws/rl"
	maxminddb "github.com/oschwald/maxminddb-golang"
)

type CountryLimiter struct {
	db                         *maxminddb.Reader
	limitRateForOtherCountries bool
	countries                  []string
	BaseLimiter
}

// 国別のリクエスト数を制限する
// 制限単位はIPアドレス
func NewCountryLimiter(
	dbPath string,
	countries []string,
	reqLimit int,
	windowLen time.Duration,
	targetExtensions []string,
	limitRateForOtherCountries bool, // 他の国々にレートリミットを適用する
	onRequestLimit func(*rl.Context, string) http.HandlerFunc,
) (*CountryLimiter, error) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &CountryLimiter{
		db:                         db,
		countries:                  countries,
		limitRateForOtherCountries: limitRateForOtherCountries,
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

	if country == "" {
		return &rl.Rule{ReqLimit: -1}, nil
	}

	for _, c := range l.countries {
		if country == c {
			if l.limitRateForOtherCountries {
				return &rl.Rule{ReqLimit: -1}, nil
			}
			return &rl.Rule{
				Key:       remoteAddr,
				ReqLimit:  l.reqLimit,
				WindowLen: l.windowLen,
			}, nil
		}
	}

	if l.limitRateForOtherCountries {
		return &rl.Rule{
			Key:       remoteAddr,
			ReqLimit:  l.reqLimit,
			WindowLen: l.windowLen,
		}, nil
	}
	return &rl.Rule{ReqLimit: -1}, nil
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
