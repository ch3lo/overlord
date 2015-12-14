package api

import "github.com/gin-gonic/gin"

func Routes() *gin.Engine {
	router := gin.New()

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
