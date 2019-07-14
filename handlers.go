package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	period     = 5
	timeFormat = "2006-01-02T15:04:05"
)

type m bson.M
type d bson.D

type MarketDataPairs struct {
	from   time.Time
	to     time.Time
	first  MarketData
	second MarketData
}

type MarketPairDoc struct {
	Hour       time.Time             `bson:"hour"`
	MarketPair string                `bson:"market_pair"`
	Minutes    map[string]MarketData `bson:"minutes"`
}

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

type MarketData struct {
	MarketName string    `json:"marketname"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Volume     float64   `json:"volume"`
	Created    time.Time `json:"created"`
	Timestamp  time.Time `json:"timestamp"`
}

func checkParam(r *http.Request, param string) (string, error) {
	params, ok := r.URL.Query()[param]
	if !ok || len(params[0]) < 1 {
		message := fmt.Sprintf("Can't process: Url Param '%s' is missing", param)
		return "", errors.New(message)
	}
	return params[0], nil
}

func checkTimeFormat(k string, v string) (time.Time, error) {
	vTime, err := time.Parse(timeFormat, v)
	if err != nil {
		message := fmt.Sprintf("Can't process: '%s' not in format: '2006-01-02T15:04:05'", k)
		return time.Time{}, errors.New(message)
	}
	return vTime, nil
}

func createMarketPairSummary(el MarketDataPairs) (MarketPairSummary, error) {
	var err error
	if el.first.Volume == 0 || el.first.High == 0 || el.first.Low == 0 {
		return MarketPairSummary{}, errors.New("Divide by zero")
	}
	if el.second.Volume == 0 || el.second.High == 0 || el.second.Low == 0 {
		return MarketPairSummary{}, errors.New("Missing data for comparison")
	}
	marketPairMetric := MarketPairMetric{
		MarketPair: el.first.MarketName,
		MarketPairGrowthMetric: MarketPairGrowthMetric{
			VolumeGrowth: (el.second.Volume - el.first.Volume) * 100.0 / el.first.Volume,
			HighGrowth:   (el.second.High - el.first.High) * 100.0 / el.first.High,
			LowGrowth:    (el.second.Low - el.first.Low) * 100.0 / el.first.Low,
		},
	}
	marketPairSummary := MarketPairSummary{
		From: el.from,
		To:   el.to,
		Data: []MarketPairMetric{marketPairMetric},
	}
	return marketPairSummary, err
}

func containsPairSummary(a []MarketPairSummary, x MarketPairSummary) bool {
	for i := range a {
		if a[i].From == x.From && a[i].To == x.To {
			return true
		}
	}
	return false
}

func appendSummaryFromDataPairs(b []MarketDataPairs) []MarketPairSummary {
	a := make([]MarketPairSummary, 0)
	for _, el := range b {
		marketPairSummary, err := createMarketPairSummary(el)
		if err != nil {
			glog.Error(err)
		}
		if err == nil {
			a = append(a, marketPairSummary)
		}
	}
	return a
}

func mapifySummaries(s []MarketPairSummary) ([]string, map[string][]MarketPairMetric) {
	a := make([]string, 0)
	m := make(map[string][]MarketPairMetric)
	for _, el := range s {
		// fmt.Printf("%+v\n", el)
		key := fmt.Sprintf("%s TO %s", el.From.Format(timeFormat), el.To.Format(timeFormat))
		a = append(a, key)
		var data MarketPairMetric
		if len(el.Data) == 0 {
			data = MarketPairMetric{}
		} else {
			data = el.Data[0]
		}
		m[key] = append(m[key], data)
	}
	return a, m
}

func prepareResponse(a []string, m map[string][]MarketPairMetric) []MarketPairSummary {
	s := make([]MarketPairSummary, 0)
	for i, el := range a {
		if i == len(m) {
			break
		}
		t := strings.Split(el, " TO ")
		t1, _ := time.Parse(timeFormat, t[0])
		t2, _ := time.Parse(timeFormat, t[1])
		// fmt.Printf("%+v\n", m[el])
		m[el] = m[el][1:]
		marketPairSummary := MarketPairSummary{
			From: t1,
			To:   t2,
			Data: m[el],
		}
		s = append(s, marketPairSummary)
	}
	return s
}

func exportAnalyticsEndpoint(w http.ResponseWriter, r *http.Request) {

	format, err := checkParam(r, "format")
	if err != nil {
		glog.Error(err.Error())
		http.Error(w, err.Error(), 422)
		return
	}
	if format != "json" {
		message := "Can't process: Unsupported format. Supported format(s): json."
		glog.Error(message)
		http.Error(w, message, 422)
		return
	}

	from, err := checkParam(r, "from")
	if err != nil {
		glog.Error(err.Error())
		http.Error(w, err.Error(), 422)
		return
	}
	fromTime, err := checkTimeFormat("from", from)
	if err != nil {
		glog.Info(err.Error())
		http.Error(w, err.Error(), 422)
		return
	}

	to, err := checkParam(r, "to")
	if err != nil {
		glog.Error(err.Error())
		http.Error(w, err.Error(), 422)
		return
	}
	toTime, err := checkTimeFormat("to", to)
	if err != nil {
		glog.Error(err.Error())
		http.Error(w, err.Error(), 422)
		return
	}

	if fromTime.After(toTime) {
		message := "Can't process: 'from' is after 'to'."
		glog.Error(message)
		http.Error(w, message, 422)
		return
	}

	fromTimeHour := fromTime.Truncate(time.Hour)
	toTimeHour := toTime.Truncate(time.Hour)
	fromTimeMin := fromTime.Truncate(time.Minute)
	toTimeMin := toTime.Truncate(time.Minute)

	t1 := fromTimeMin.Minute()
	t2 := 59
	carryOver := 0

	marketDataPairsMap := make(map[string]*MarketDataPairs)
	marketDataPairsArray := make([]MarketDataPairs, 0)
	marketPairSummaries := make([]MarketPairSummary, 0)

	collection := client.Database("cryptocurrencies").Collection("marketvalues")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := GetFromToFilterBSON(fromTimeHour, toTimeHour)

	findOptions := options.Find()
	findOptions.SetSort(map[string]int{"hour": 1})

	cur, err := collection.Find(ctx, filter, findOptions)

	if err != nil {
		glog.Info(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), 500)
	}

	p := MarketPairSummary{}
	for i := fromTimeMin; i.Before(toTimeMin) || i.Equal(toTimeMin); i = i.Add(time.Duration(5) * time.Minute) {
		if i == fromTimeMin {
			p.From = i
		} else {
			p.To = i
			if !containsPairSummary(marketPairSummaries, p) {
				marketPairSummaries = append(marketPairSummaries, p)
			}
			p = MarketPairSummary{
				From: i,
			}
		}
	}

	for cur.Next(context.TODO()) {
		var doc MarketPairDoc
		err := cur.Decode(&doc)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), 500)
		}
		if _, ok := marketDataPairsMap[doc.MarketPair]; !ok {
			glog.Info("setting pair in map")
			glog.Info("map state: ", marketDataPairsMap)
			marketDataPairsMap[doc.MarketPair] = &MarketDataPairs{}
		}
		// for the next batch of docs, when the hour has flipped over
		if doc.Hour != fromTimeHour {
			fromTimeHour = doc.Hour
			t1 = carryOver
		}
		// last minute as per request
		if doc.Hour == toTimeHour {
			t2 = toTimeMin.Minute()
		}
		glog.Info("t1 ", t1)
		glog.Info("t2 ", t2)
		for i := t1; i <= t2; i += 5 {
			glog.Info("i ", i)
			// to manage minute roll over / carry over to next hour
			carryOver = 5 - (t2 - i) - 1
			m := marketDataPairsMap[doc.MarketPair]
			glog.Info("in map:", *m)
			marketData := doc.Minutes[strconv.Itoa(i)]

			if i == t1 && (*m).from.Equal(MarketDataPairs{}.from) {
				(*m).from = doc.Hour.Add(time.Duration(i) * time.Minute)
				(*m).first = marketData
			} else {
				(*m).to = doc.Hour.Add(time.Duration(i) * time.Minute)
				(*m).second = marketData

				marketDataPairsArray = append(marketDataPairsArray, *m)
				// fmt.Printf("%+v\n", *m)
				marketDataPairsMap[doc.MarketPair] = &MarketDataPairs{
					from:   (*m).to,
					first:  (*m).second,
					to:     time.Time{},
					second: MarketData{},
				}
			}
		}

	}
	// Close the cursor once finished
	cur.Close(context.TODO())

	// performaing growth calculations
	calculatedMarketPairSummaries := appendSummaryFromDataPairs(marketDataPairsArray)

	// add growth calculations
	for _, el := range calculatedMarketPairSummaries {
		marketPairSummaries = append(marketPairSummaries, el)
	}

	// collecting into same periods
	a, m := mapifySummaries(marketPairSummaries)

	// creating the response data
	s := prepareResponse(a, m)

	// returning JSON back
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
