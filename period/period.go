package period

import (
	"errors"
	"time"

	m "github.com/steotia/go-analytics-crypto-api/marketdata"
	"github.com/steotia/go-analytics-crypto-api/metrics"
)

type Periods struct {
	PeriodSlots []*PeriodSlot
}

func (p *Periods) AddPeriod(s *PeriodSlot) {
	p.PeriodSlots = append(p.PeriodSlots, s)
}

func (p *Periods) SetMarketData(d m.MarketData) {
	truncatedTS := d.Timestamp.Truncate(time.Minute)
	setNext := false
	for i, s := range p.PeriodSlots {
		(*s).resetMarketDataPairsMap(d.MarketName)
		if s.From == truncatedTS {
			(*s.MarketDataPairsMap[d.MarketName]).setLeft(d)
		}
		if s.To == truncatedTS {
			(*s.MarketDataPairsMap[d.MarketName]).setRight(d)
			setNext = true
		}
		if truncatedTS.Before(s.From) {
			(*s.MarketDataPairsMap[d.MarketName]).setLeft(d)
			(*s.MarketDataPairsMap[d.MarketName]).setRight(d)
		}
		if truncatedTS.After(s.From) && truncatedTS.Before(s.To) {
			(*s.MarketDataPairsMap[d.MarketName]).setRight(d)
			setNext = true
		}
		if setNext {
			if i < len(p.PeriodSlots)-1 {
				nexts := p.PeriodSlots[i+1]
				(*nexts).resetMarketDataPairsMap(d.MarketName)
				(*nexts.MarketDataPairsMap[d.MarketName]).setLeft(d)
			}
			setNext = false
		}
	}
}

func (p *Periods) GenerateMetrics() []metrics.MarketPairSummary {
	marketPairSummaries := make([]metrics.MarketPairSummary, 0)
	for _, s := range p.PeriodSlots {
		marketPairSummary := metrics.MarketPairSummary{
			From: s.From,
			To:   s.To,
			Data: []metrics.MarketPairMetric{},
		}
		for _, v := range s.MarketDataPairsMap {
			if v.Left.Volume == 0 || v.Left.High == 0 || v.Left.Low == 0 {
				continue
			}
			if v.Right.Volume == 0 || v.Right.High == 0 || v.Right.Low == 0 {
				continue
			}
			marketPairMetric := metrics.MarketPairMetric{
				MarketPair: v.Left.MarketName,
				MarketPairGrowthMetric: metrics.MarketPairGrowthMetric{
					VolumeGrowth: rate(v.Left.Volume, v.Right.Volume),
					HighGrowth:   rate(v.Left.High, v.Right.High),
					LowGrowth:    rate(v.Left.Low, v.Right.Low),
				},
			}
			marketPairSummary.Data = append(marketPairSummary.Data, marketPairMetric)
		}
		marketPairSummaries = append(marketPairSummaries, marketPairSummary)
	}
	return marketPairSummaries
}

// todo
func rate(a float64, b float64) float64 {
	return (b - a) * 100.0 / a
}

type PeriodSlot struct {
	From               time.Time
	To                 time.Time
	MarketDataPairsMap map[string]*MarketDataPair
}

func (p *PeriodSlot) resetMarketDataPairsMap(s string) {
	if _, ok := p.MarketDataPairsMap[s]; !ok {
		p.MarketDataPairsMap[s] = &MarketDataPair{}
	}
}

type MarketDataPair struct {
	Left  m.MarketData `json:"left"`
	Right m.MarketData `json:"right"`
}

func (m *MarketDataPair) setLeft(a m.MarketData) {
	m.Left = a
}

func (m *MarketDataPair) setRight(b m.MarketData) {
	m.Right = b
}

func NewBlankPeriod(a time.Time, b time.Time) *PeriodSlot {
	return &PeriodSlot{
		From:               a.Truncate(time.Minute),
		To:                 b.Truncate(time.Minute),
		MarketDataPairsMap: map[string]*MarketDataPair{},
	}
}

func NewBlankPeriodsBetween(a time.Time, b time.Time, d time.Duration) (Periods, error) {
	a = a.Truncate(time.Minute)
	b = b.Truncate(time.Minute)
	if !a.Before(b) {
		return Periods{}, errors.New("'from' should be less than 'to'")
	}
	periods := Periods{}
	for i := a; !i.After(b) && !i.Equal(b); {
		nexti := i.Add(d)
		p := NewBlankPeriod(i, nexti)
		if !nexti.Before(b) {
			p = NewBlankPeriod(i, b)
		}
		periods.AddPeriod(p)
		i = nexti
	}
	return periods, nil
}
