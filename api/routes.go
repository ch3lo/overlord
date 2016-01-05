package api

import (
	"github.com/gorilla/mux"
	"github.com/thoas/stats"
)

var routesMap = map[string]map[string]serviceHandler{
	"GET": {
		"/":                       getServices,
		"/{service_id}":           getServiceByServiceId,
		"/{service_id}/{cluster}": getServiceByClusterAndServiceId,
	},
	"PUT": {
		"/": putService,
		"/{service_id}/versions": putServiceVersionByServiceId,
	},
}

func routes(sts *stats.Stats) *mux.Router {
	router := mux.NewRouter()

	sh := &statsHandler{s: sts}
	router.Handle("/stats", sh).Methods("GET")

	// API v1
	v1Services := router.PathPrefix("/api/v1/services").Subrouter()

	for method, mappings := range routesMap {
		for path, h := range mappings {
			v1Services.Handle(path, &errorHandler{h}).Methods(method)
		}
	}

	return router
}
