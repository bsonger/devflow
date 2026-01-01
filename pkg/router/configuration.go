package router

import (
	"github.com/bsonger/devflow/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterConfigurationRoutes(rg *gin.RouterGroup) {
	cfg := rg.Group("/configurations")

	cfg.GET("", api.ConfigurationRouteApi.List)
	cfg.GET("/:id", api.ConfigurationRouteApi.Get)
	cfg.POST("", api.ConfigurationRouteApi.Create)
	cfg.PUT("/:id", api.ConfigurationRouteApi.Update)
	cfg.DELETE("/:id", api.ConfigurationRouteApi.Delete)
}
