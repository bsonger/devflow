package router

import (
	"context"
	"github.com/bsonger/devflow-common/client/logging"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	_ "github.com/bsonger/devflow/docs" // swagger docs 自动生成
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func shouldIgnore(path string) bool {
	return path == "/metrics" || path == "/health" || strings.HasPrefix(path, "/swagger")
}

func GinMetricsMiddleware() gin.HandlerFunc {
	meter := otel.Meter("devflow/http")
	requestsCounter, _ := meter.Int64Counter("http.server.requests")
	requestLatency, _ := meter.Float64Histogram(
		"http.server.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(
			0.3, 0.5, 1, 3, // 👈 更少的桶（重点）
		),
	)

	requestSize, _ := meter.Int64Histogram(
		"http.server.request.size",
		metric.WithUnit("By"),
		metric.WithExplicitBucketBoundaries(
			512, 2048, 8192, 32768,
		),
	)

	responseSize, _ := meter.Int64Histogram(
		"http.server.response.size",
		metric.WithUnit("By"),
		metric.WithExplicitBucketBoundaries(
			512, 2048, 8192, 32768,
		),
	)

	const highLatencyThreshold = 1.0 // 秒

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if shouldIgnore(c.Request.URL.Path) {
			return
		}
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		ctx := c.Request.Context()
		attrs := []attribute.KeyValue{
			attribute.String("method", c.Request.Method),
			attribute.String("path", c.FullPath()),
			attribute.Int("status", status),
		}

		// request size
		reqSize := c.Request.ContentLength
		if reqSize < 0 {
			reqSize = 0
		}

		if status >= 500 || duration > highLatencyThreshold {
			requestsCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
			requestLatency.Record(ctx, duration, metric.WithAttributes(attrs...))
		} else {
			noopCtx := context.Background()
			requestsCounter.Add(noopCtx, 1, metric.WithAttributes(attrs...))
			requestLatency.Record(noopCtx, duration, metric.WithAttributes(attrs...))
		}
		requestSize.Record(ctx, reqSize, metric.WithAttributes(attrs...))
		responseSize.Record(ctx, int64(c.Writer.Size()), metric.WithAttributes(attrs...))

	}
}

// GinZapLogger 返回一个 Gin 中间件，将日志写入 zap
func GinZapLogger(loggerFunc func(ctx context.Context) *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		loggerFunc(c).Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Duration("latency", latency),
			zap.String("error", errorMessage),
		)
	}
}

func GinZapRecovery(loggerFunc func(ctx context.Context) *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				// 打印 panic 到 zap
				loggerFunc(c).Error("panic recovered",
					zap.Any("panic", rec),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
				)
				// 返回 500 给客户端
				c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

// NewRouter creates the main Gin router.
func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New() // ⭐ 不使用 gin.Default()

	var myFilter otelgin.Filter = func(req *http.Request) bool {
		path := req.URL.Path
		if path == "/metrics" || strings.HasPrefix(path, "/swagger") {
			return false
		}
		return true
	}

	r.Use(otelgin.Middleware("devflow", otelgin.WithFilter(myFilter)))

	r.Use(GinMetricsMiddleware())
	r.Use(GinZapLogger(logging.LoggerWithContext), GinZapRecovery(logging.LoggerWithContext))
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
