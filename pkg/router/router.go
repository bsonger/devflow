package router

import (
	_ "github.com/bsonger/devflow/docs" // swagger docs 自动生成
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"net/http"
	"time"
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
		otelgin.Middleware("devflow", otelgin.WithFilter(myFilter)),
		LoggerMiddleware(),
		GinZapRecovery(),
		//PyroscopeMiddleware(),
		//GinMetricsMiddleware(),
		GinZapLogger(),
		cors.New(cors.Config{
			AllowOrigins:     []string{"*"}, // 允许所有来源
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
	)

	// 1️⃣ Swagger UI 路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 2️⃣ API 分组
	api := r.Group("/api/v1")

	// 3️⃣ 注册 Application 路由
	RegisterApplicationRoutes(api)
	RegisterManifestRoutes(api)
	RegisterJobRoutes(api)
	return r
}

func StartMetricsServer(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			panic(err)
		}
	}()
}
