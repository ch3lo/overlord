package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
)

func Routes() *gin.Engine {
	router := gin.New()

	router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "POST, GET, OPTIONS, PUT, DELETE, UPDATE",
		RequestHeaders:  "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
		ExposedHeaders:  "Content-Length",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	// API v1
	v1Services := router.Group("/api/v1/services")
	v1Services.GET("/", GetServices)
	v1Services.PUT("/", PutService)
	//v1Services.GET("/test", ServicesTestGet)

	v1Services.GET("/:service_id", GetServiceByServiceId)
	v1Services.PUT("/:service_id/versions", PutServiceVersionByServiceId)

	v1Services.GET("/:service_id/:cluster", GetServiceByClusterAndServiceId)

	return router
}
