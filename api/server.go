package api

import (
	"github.com/ch3lo/overlord/configuration"
	"github.com/codegangsta/negroni"
	"github.com/rs/cors"
	"github.com/thoas/stats"
)

func Server(config *configuration.Configuration) {
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST, GET, OPTIONS, PUT, DELETE, UPDATE"},
		AllowedHeaders:   []string{"Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization"},
		ExposedHeaders:   []string{"Content-Length"},
		MaxAge:           50,
		AllowCredentials: true,
	})

	statsMiddleware := stats.New()

	router := routes(config, statsMiddleware)

	n := negroni.Classic()
	n.Use(corsMiddleware)
	n.Use(statsMiddleware)
	n.UseHandler(router)

	n.Run(":8080")
}
