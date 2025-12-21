package router

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"time"

	_ "github.com/bsonger/devflow/docs" // swagger docs 自动生成
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func GinMetricsMiddleware() gin.HandlerFunc {
	meter := otel.Meter("devflow/http")
	requestsCounter, _ := meter.Int64Counter("http.server.requests")
	requestLatency, _ := meter.Float64Histogram("http.server.duration")

	const highLatencyThreshold = 1.0 // 秒

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		ctx := c.Request.Context()
		attrs := []attribute.KeyValue{
			attribute.String("method", c.Request.Method),
			attribute.String("path", c.FullPath()),
			attribute.Int("status", status),
		}

		if status >= 500 || duration > highLatencyThreshold {
			requestsCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
			requestLatency.Record(ctx, duration, metric.WithAttributes(attrs...))
		} else {
			noopCtx := context.Background()
			requestsCounter.Add(noopCtx, 1, metric.WithAttributes(attrs...))
			requestLatency.Record(noopCtx, duration, metric.WithAttributes(attrs...))
		}
	}
}

// NewRouter creates the main Gin router.
func NewRouter() *gin.Engine {
	r := gin.Default()

	r.Use(otelgin.Middleware("devflow"), GinMetricsMiddleware())
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
