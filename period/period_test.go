package period

import (
	"reflect"
	"testing"
	"time"

	"github.com/steotia/go-analytics-crypto-api/marketdata"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	t2, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:18:41+00:00")
	period := NewBlankPeriod(t1, t2)
	expected := PeriodSlot{
		From:               time.Time{},
		To:                 time.Time{},
		MarketDataPairsMap: map[string]*MarketDataPair{},
	}
	if reflect.DeepEqual(*period, expected) {
		t.Errorf("expected '%v' but got '%v'", expected, *period)
	}
}

func TestValidatePeriodGeneratorForInvalidInputs(t *testing.T) {
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	gap := time.Duration(5) * time.Minute
	ps, err := NewBlankPeriodsBetween(t1, t1, gap)
	assert.NotNil(t, err)
	ps, err = NewBlankPeriodsBetween(t1.Add(gap), t1, gap)
	assert.NotNil(t, err)
	ps, err = NewBlankPeriodsBetween(t1, t1.Add(gap), gap)
	assert.Nil(t, err)
	assert.NotNil(t, ps)
}

func TestPeriodGeneratorWithLargeGap(t *testing.T) {
	largegap := time.Duration(5) * time.Minute
	lessgap := time.Duration(4) * time.Minute
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	t2 := t1.Add(lessgap)
	ps, err := NewBlankPeriodsBetween(t1, t2, largegap)
	assert.Nil(t, err)
	assert.Len(t, ps.PeriodSlots, 1)
	expected := NewBlankPeriod(t1, t2)
	assert.Equal(t, expected, ps.PeriodSlots[0])
}

func TestPeriodGeneratorWithExactGap(t *testing.T) {
	gap := time.Duration(5) * time.Minute
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	t2 := t1.Add(gap)
	ps, err := NewBlankPeriodsBetween(t1, t2, gap)
	assert.Nil(t, err)
	assert.Len(t, ps.PeriodSlots, 1)
	expected := NewBlankPeriod(t1, t2)
	assert.Equal(t, expected, ps.PeriodSlots[0])
}

func TestPeriodGeneratorWithSmallGap(t *testing.T) {

	// var buf bytes.Buffer
	// log.SetOutput(&buf)
	// defer func() {
	// 	log.SetOutput(os.Stderr)
	// }()

	largegap := time.Duration(9) * time.Minute
	lessgap := time.Duration(4) * time.Minute
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	t2 := t1.Add(largegap)
	ps, err := NewBlankPeriodsBetween(t1, t2, lessgap)
	assert.Nil(t, err)
	assert.Len(t, ps.PeriodSlots, 3)
	x1 := t1.Add(lessgap)
	x2 := x1.Add(lessgap)
	e1 := NewBlankPeriod(t1, x1)
	e2 := NewBlankPeriod(x1, x2)
	e3 := NewBlankPeriod(x2, t2)
	assert.Equal(t, *e1, *ps.PeriodSlots[0])
	assert.Equal(t, *e2, *ps.PeriodSlots[1])
	assert.Equal(t, *e3, *ps.PeriodSlots[2])

	// t.Log(buf.String())
}

func TestPeriodSlotPopulateWithMarketDataOutsidePeriod(t *testing.T) {

	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	gap := time.Duration(5) * time.Minute
	t2 := t1.Add(gap)

	periods, _ := NewBlankPeriodsBetween(t1, t2, gap)

	mData := marketdata.MarketData{
		High:      1,
		Timestamp: time.Time{},
	}

	slot := periods.PeriodSlots[0]

	val, ok := slot.MarketDataPairsMap[mData.MarketName]
	assert.False(t, ok)

	periods.SetMarketData(mData)

	val, ok = slot.MarketDataPairsMap[mData.MarketName]
	assert.True(t, ok)
	assert.Equal(t, mData, val.Left)
	assert.Equal(t, mData, val.Right)
}

func TestPeriodSlotPopulateWithMarketDataExactPeriod(t *testing.T) {
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	gap := time.Duration(5) * time.Minute
	t2 := t1.Add(gap)

	periods, _ := NewBlankPeriodsBetween(t1, t2, gap)

	mData := marketdata.MarketData{
		High:      1,
		Timestamp: time.Time{},
	}
	slot := periods.PeriodSlots[0]
	mData.Timestamp = t1
	periods.SetMarketData(mData)

	val, _ := slot.MarketDataPairsMap[mData.MarketName]
	assert.Equal(t, mData, val.Left)
	assert.Equal(t, marketdata.MarketData{}, val.Right)

	mData2 := marketdata.MarketData{
		High:      2,
		Timestamp: t2,
	}
	periods.SetMarketData(mData2)
	assert.Equal(t, mData, val.Left)
	assert.Equal(t, mData2, val.Right)

	// assert.False(t, set)
	// assert.Empty(t, period.MarketDataPairs)
}

func TestPeriodSlotPopulateWithMarketDataInBetweenPeriod(t *testing.T) {
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	gap := time.Duration(5) * time.Minute
	lessthangap := time.Duration(3) * time.Minute
	t2 := t1.Add(gap)

	periods, _ := NewBlankPeriodsBetween(t1, t2, gap)
	slot := periods.PeriodSlots[0]

	mData := marketdata.MarketData{
		High:      1,
		Timestamp: t1.Add(lessthangap),
	}
	periods.SetMarketData(mData)

	val, _ := slot.MarketDataPairsMap[mData.MarketName]
	assert.Equal(t, marketdata.MarketData{}, val.Left)
	assert.Equal(t, mData, val.Right)
}

func TestPeriodsNeighbouringSlotWithMarketData(t *testing.T) {
	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")
	gap := time.Duration(5) * time.Minute
	sometime := time.Duration(13) * time.Minute
	t2 := t1.Add(5 * gap)

	periods, _ := NewBlankPeriodsBetween(t1, t2, gap)
	mData := marketdata.MarketData{
		High:      1,
		Timestamp: t1.Add(sometime),
	}
	periods.SetMarketData(mData)
	prevslot := periods.PeriodSlots[1]
	slot := periods.PeriodSlots[2]
	nextslot := periods.PeriodSlots[3]
	nextnextslot := periods.PeriodSlots[4]
	prevval, _ := prevslot.MarketDataPairsMap[mData.MarketName]
	val, _ := slot.MarketDataPairsMap[mData.MarketName]
	nextval, _ := nextslot.MarketDataPairsMap[mData.MarketName]
	nextnextval, _ := nextnextslot.MarketDataPairsMap[mData.MarketName]
	assert.Equal(t, marketdata.MarketData{}, prevval.Left)
	assert.Equal(t, marketdata.MarketData{}, prevval.Right)
	assert.Equal(t, marketdata.MarketData{}, val.Left)
	assert.Equal(t, mData, val.Right)
	assert.Equal(t, mData, nextval.Left)
	assert.Equal(t, mData, nextval.Right)
	assert.Equal(t, mData, nextnextval.Left)
	assert.Equal(t, mData, nextnextval.Right)
}

func TestPeriodsGenerateMetrics(t *testing.T) {