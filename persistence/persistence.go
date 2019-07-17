package persistence

import (
	"time"

	"github.com/steotia/go-analytics-crypto-api/marketdata"
)

type MarketPairDoc struct {
	Hour       time.Time                        `bson:"hour"`
	MarketPair string                           `bson:"market_pair"`
	Minutes    map[string]marketdata.MarketData `bson:"minutes"`
}

type Next interface {
	Next()
	Close()
}
type TimeSeries struct {
}
