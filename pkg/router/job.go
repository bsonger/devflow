package router

import (
	"github.com/bsonger/devflow/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterJobRoutes(rg *gin.RouterGroup) {
	job := rg.Group("/jobs")

	job.GET("", api.JobRouteApi.List)
	job.GET("/:id", api.JobRouteApi.Get)
	job.POST("", api.JobRouteApi.Create)
}
