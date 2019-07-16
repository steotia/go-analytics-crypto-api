package marketdata

import "time"

type MarketData struct {
	MarketName string    `json:"marketname"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Volume     float64   `json:"volume"`
	Created    time.Time `json:"created"`
	Timestamp  time.Time `json:"timestamp"`
}
