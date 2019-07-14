package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GetFromToFilterBSON(fromTimeHour time.Time, toTimeHour time.Time) bson.M {
	return bson.M{
		"$and": bson.A{
			bson.M{
				"hour": bson.M{
					"$gte": fromTimeHour,
				},
			},
			bson.M{
				"hour": bson.M{
					"$lte": toTimeHour,
				},
			},
			// bson.M{
			// 	"market_pair": bson.M{
			// 		"$eq": "BTC-ETH",
			// 	},
			// },
		},
	}
}
