package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/steotia/go-analytics-crypto-api/marketdata"
	"github.com/steotia/go-analytics-crypto-api/period"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	timeFormat = "2006-01-02T15:04:05"
)

type m bson.M
type d bson.D

type MarketDataPairs struct {
	from   time.Time
	to     time.Time
	first  marketdata.MarketData
	second marketdata.MarketData
}

type MarketPairDoc struct {
	Hour       time.Time                        `bson:"hour"`
	MarketPair string                           `bson:"market_pair"`
	Minutes    map[string]marketdata.MarketData `bson:"minutes"`
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
		return
	}
	gap := time.Duration(5) * time.Minute
	periods, err := period.NewBlankPeriodsBetween(fromTimeMin, toTimeMin, gap)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), 500)
		return
	}

	for cur.Next(context.TODO()) {
		var doc MarketPairDoc
		err := cur.Decode(&doc)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), 500)
		}
		for _, data := range doc.Minutes {
			periods.SetMarketData(data)
		}
	}
	cur.Close(context.TODO())

	// returning JSON back
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(periods.GenerateMetrics())
}
