package router

import (
	"github.com/bsonger/devflow/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterApplicationRoutes(rg *gin.RouterGroup) {
	app := rg.Group("/applications")

	app.GET("", api.ApplicationRouteApi.List)
	app.GET("/:id", api.ApplicationRouteApi.Get)
	app.POST("", api.ApplicationRouteApi.Create)
	app.PUT("/:id", api.ApplicationRouteApi.Update)
	app.DELETE("/:id", api.ApplicationRouteApi.Delete)
	app.PATCH("/:id/active_manifest", api.ApplicationRouteApi.UpdateActiveManifest)

	RegisterManifestRoutes(app)
}
