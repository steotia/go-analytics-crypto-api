package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/steotia/go-analytics-crypto-api/period"
	"github.com/steotia/go-analytics-crypto-api/persistence"
	"github.com/steotia/go-analytics-crypto-api/validators"
)

func exportAnalyticsEndpoint(w http.ResponseWriter, r *http.Request) {

	getMetricsInput, err := validators.NewGetMetricsInput(r)
	if err != nil {
		glog.Error(err.Error())
		http.Error(w, err.Error(), 422)
		return
	}
	err = getMetricsInput.Validate()
	if err != nil {
		glog.Error(err.Error())
		http.Error(w, err.Error(), 422)
		return
	}

	fromTimeHour := getMetricsInput.FromTime.Truncate(time.Hour)
	toTimeHour := getMetricsInput.ToTime.Truncate(time.Hour)
	fromTimeMin := getMetricsInput.FromTime.Truncate(time.Minute)
	toTimeMin := getMetricsInput.ToTime.Truncate(time.Minute)
	/*
		client, _ := persistence.GetMongoDBClient()
		// if err != nil {
		// 	return MongoTimeSeries{}, nil
		// }
		collection := client.Database("cryptocurrencies").Collection("marketvalues")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		filter := persistence.GetFromToFilterBSON(fromTimeHour, toTimeHour)

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
	*/
	gap := time.Duration(5) * time.Minute
	periods, err := period.NewBlankPeriodsBetween(fromTimeMin, toTimeMin, gap)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), 500)
		return
	}

	mongoTimeSeries, err := persistence.NewMongoTimeSeries(fromTimeHour, toTimeHour)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), 500)
	}
	glog.Info("======>")
	for mongoTimeSeries.Next() {
		var doc persistence.MarketPairDoc
		err := mongoTimeSeries.Decode(&doc)
		glog.Info(doc)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), 500)
		}
		for _, data := range doc.Minutes {
			periods.SetMarketData(data)
		}
	}
	mongoTimeSeries.Close()
	glog.Info("<======")

	/*
		for cur.Next(context.TODO()) {
			var doc persistence.MarketPairDoc
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
	*/
	// // returning JSON back
	w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(periods.GenerateMetrics())
	json.NewEncoder(w).Encode(periods)
}
