package api

import (
	"github.com/ch3lo/overlord/configuration"
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

func routes(config *configuration.Configuration, sts *stats.Stats) *mux.Router {
	ctx := newContext(config)
	router := mux.NewRouter()

	router.Handle("/stats", &statsHandler{sts}).Methods("GET")

	// API v1
	v1Services := router.PathPrefix("/api/v1/services").Subrouter()

	for method, mappings := range routesMap {
		for path, h := range mappings {
			v1Services.Handle(path, errorHandler{h, ctx}).Methods(method)
		}
	}

	return router
}
