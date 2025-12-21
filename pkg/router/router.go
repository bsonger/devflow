package router

import (
	"github.com/bsonger/devflow-common/client/logging"
	_ "github.com/bsonger/devflow/docs" // swagger docs 自动生成
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"net/http"
)

// NewRouter creates the main Gin router.
func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New() // ⭐ 不使用 gin.Default()

	var myFilter otelgin.Filter = func(req *http.Request) bool {
		path := req.URL.Path
		return !shouldIgnore(path)
	}

	r.Use(
		GinZapRecovery(logging.LoggerWithContext), // ⭐ 最前
		otelgin.Middleware("devflow", otelgin.WithFilter(myFilter)),
		//GinMetricsMiddleware(),
		GinZapLogger(logging.LoggerWithContext),
	)

	// 1️⃣ Swagger UI 路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 2️⃣ API 分组
	api := r.Group("/api/v1")

	// 3️⃣ 注册 Application 路由
	RegisterApplicationRoutes(api)
	RegisterManifestRoutes(api)
	RegisterJobRoutes(api)
	//service.StartTektonInformer(context.Background())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return r
}
