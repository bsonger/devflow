package router

import (
	"github.com/bsonger/devflow/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterJobRoutes(rg *gin.RouterGroup) {
	app := rg.Group("/jobs")

	app.POST("", api.JobRouteApi.Create)
	app.GET("/:id", api.JobRouteApi.Get)
	//app.PUT("/:id", api.JobRouteApi.Update)
	//app.DELETE("/:id", api.JobRouteApi.Delete)
	app.GET("/", api.JobRouteApi.List)
}
