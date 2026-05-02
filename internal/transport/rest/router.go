package rest

import (
	"github.com/gin-gonic/gin"
)

// NewRouter creates a configured gin.Engine with all routes registered.
func NewRouter(handler *Handler, apiKey string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	v1.Use(APIKeyMiddleware(apiKey))
	{
		v1.POST("/deploy", handler.Deploy)
		v1.GET("/stats", handler.Stats)
	}

	return r
}
