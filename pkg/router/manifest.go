package router

import (
	"github.com/bsonger/devflow/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterManifestRoutes(rg *gin.RouterGroup) {

	rg.POST("/manifests", api.ManifestRouteApi.Create)
}
