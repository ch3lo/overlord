package api

import (
	"github.com/gorilla/mux"
	"github.com/thoas/stats"
)

func Routes(sts *stats.Stats) *mux.Router {
	router := mux.NewRouter()

	sh := &statsHandler{s: sts}
	router.Handle("/stats", sh).Methods("GET")

	// API v1
	v1Services := router.PathPrefix("/api/v1/services").Subrouter()
	v1Services.HandleFunc("/", getServices).Methods("GET")
	v1Services.HandleFunc("/", putService).Methods("PUT")
	//v1Services.GET("/test", ServicesTestGet)

	v1Services.HandleFunc("/{service_id}", getServiceByServiceId).Methods("GET")
	v1Services.HandleFunc("/{service_id}/versions", putServiceVersionByServiceId).Methods("PUT")

	v1Services.HandleFunc("/{service_id}/{cluster}", getServiceByClusterAndServiceId).Methods("GET")

	return router
}
