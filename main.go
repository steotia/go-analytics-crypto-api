package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"

	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func init() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	var err error
	client, err = GetMongoDBClient()
	if err != nil {
		glog.Fatal(err)
	}
}
func main() {
	glog.Info("Starting the application...")
	router := NewRouter()
	err := http.ListenAndServe(":12345", router)
	if err != nil {
		glog.Fatal(err)
	}
}
