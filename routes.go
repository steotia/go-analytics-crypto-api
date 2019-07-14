package main

import "github.com/gorilla/mux"

func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/export/analytics", exportAnalyticsEndpoint).Methods("GET")
	return router
}
