package persistence

import (
	"context"
	"errors"
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

type MongoTimeSeries struct {
	cancelF context.CancelFunc
	cursor  *mongo.Cursor
}

func (m *MongoTimeSeries) Next() bool {
	if m.cursor == nil {
		return false
	}
	return m.cursor.Next(context.TODO())
}

func (m *MongoTimeSeries) Decode(val interface{}) error {
	if m.cursor == nil {
		return errors.New("No cursor")
	}
	return m.cursor.Decode(&val)
}

func (m *MongoTimeSeries) Close() {
	if m.cancelF != nil {
		m.cancelF()
	}
	if m.cursor != nil {
		m.cursor.Close(context.TODO())
	}
}

func NewMongoTimeSeries(a time.Time, b time.Time) (MongoTimeSeries, error) {
	mongoTimeSeries := MongoTimeSeries{}
	client, err := GetMongoDBClient()
	if err != nil {
		return MongoTimeSeries{}, nil
	}
	collection := client.Database("cryptocurrencies").Collection("marketvalues")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	filter := GetFromToFilterBSON(a, b)
	findOptions := options.Find()
	findOptions.SetSort(map[string]int{"hour": 1})

	cur, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return mongoTimeSeries, err
	}
	mongoTimeSeries.cursor = cur
	mongoTimeSeries.cancelF = cancel
	// glog.Info("===----===>")
	// for cur.Next(context.TODO()) {
	// 	var doc MarketPairDoc
	// 	err := cur.Decode(&doc)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	for _, data := range doc.Minutes {
	// 		glog.Info(data)
	// 	}
	// }
	// cur.Close(context.TODO())
	// glog.Info("===----===>")
	return mongoTimeSeries, nil
}
