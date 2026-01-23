package router

import (
	"github.com/bsonger/devflow/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterManifestRoutes(rg *gin.RouterGroup) {
	manifest := rg.Group("/manifests")

	manifest.GET("", api.ManifestRouteApi.List)
	manifest.GET("/:id", api.ManifestRouteApi.Get)
	manifest.POST("", api.ManifestRouteApi.Create)
	manifest.PATCH("/:id", api.ManifestRouteApi.Patch)
}
