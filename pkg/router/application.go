package router

import (
	"github.com/bsonger/devflow/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterApplicationRoutes(rg *gin.RouterGroup) {
	app := rg.Group("/applications")

	app.POST("", api.ApplicationRouteApi.Create)
	app.GET("/:id", api.ApplicationRouteApi.Get)
	app.PUT("/:id", api.ApplicationRouteApi.Update)
	app.DELETE("/:id", api.ApplicationRouteApi.Delete)
	app.GET("/", api.ApplicationRouteApi.List)
	RegisterManifestRoutes(app)
}
