package persistence

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func GetMongoDBClient() (*mongo.Client, error) {
	// clientOptions := options.Client().ApplyURI("mongodb://localhost:27100")
	clientOptions := options.Client().ApplyURI("mongodb://mongo:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return client, err
}
