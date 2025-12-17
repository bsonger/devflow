package router

import (
	"context"
	"github.com/bsonger/devflow/pkg/service"
	"github.com/gin-gonic/gin"

	_ "github.com/bsonger/devflow/docs" // swagger docs 自动生成
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter creates the main Gin router.
func NewRouter() *gin.Engine {
	r := gin.Default()

	// 1️⃣ Swagger UI 路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 2️⃣ API 分组
	api := r.Group("/api/v1")

	// 3️⃣ 注册 Application 路由
	RegisterApplicationRoutes(api)
	RegisterManifestRoutes(api)
	RegisterJobRoutes(api)
	service.StartTektonInformer(context.Background())
	return r
}
