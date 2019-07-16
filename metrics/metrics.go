package metrics

import "time"

type MarketPairGrowthMetric struct {
	VolumeGrowth float64 `json:"volume_growth"`
	HighGrowth   float64 `json:"high_growth"`
	LowGrowth    float64 `json:"low_growth"`
}
type MarketPairMetric struct {
	MarketPair             string `json:"market_pair"`
	MarketPairGrowthMetric `json:"metric"`
}
type MarketPairSummary struct {
	From time.Time          `json:"from"`
	To   time.Time          `json:"to"`
	Data []MarketPairMetric `json:"market_pair_data"`
}
