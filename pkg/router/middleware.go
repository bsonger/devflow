package router

import (
	"context"
	"github.com/bsonger/devflow-common/client/logging"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grafana/pyroscope-go"
	"go.uber.org/zap"
)

/********************
 * 通用工具
 ********************/

func shouldIgnore(path string) bool {
	return path == "/metrics" ||
		path == "/health" ||
		strings.HasPrefix(path, "/swagger")
}

func routeLabel(c *gin.Context) string {
	if p := c.FullPath(); p != "" {
		return p
	}
	return "unknown"
}

/********************
 * Pyroscope Middleware
 ********************/
func PyroscopeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 request context
		ctx := c.Request.Context()

		// 获取 method 和 route（动态）
		method := c.Request.Method
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		// TagWrapper 必须包裹整个请求生命周期
		pyroscope.TagWrapper(ctx, pyroscope.Labels("http.route", route, "http.method", method), func(ctx context.Context) {
			c.Next()
		})
	}
}

/********************
 * Metrics Middleware
 ********************/
//
//func GinMetricsMiddleware() gin.HandlerFunc {
//	meter := otel.Meter("devflow")
//
//	requestsCounter, _ := meter.Int64Counter(
//		"http.server.requests",
//	)
//
//	requestLatency, _ := meter.Float64Histogram(
//		"http.server.duration",
//		metric.WithUnit("s"),
//		metric.WithExplicitBucketBoundaries(
//			0.3, 0.5, 1, 3,
//		),
//	)
//
//	requestSize, _ := meter.Int64Histogram(
//		"http.server.request.size",
//		metric.WithUnit("By"),
//		metric.WithExplicitBucketBoundaries(
//			512, 2048, 8192, 32768,
//		),
//	)
//
//	responseSize, _ := meter.Int64Histogram(
//		"http.server.response.size",
//		metric.WithUnit("By"),
//		metric.WithExplicitBucketBoundaries(
//			512, 2048, 8192, 32768,
//		),
//	)
//
//	return func(c *gin.Context) {
//		// ⭐ 必须最早过滤
//		if shouldIgnore(c.Request.URL.Path) {
//			c.Next()
//			return
//		}
//
//		start := time.Now()
//		c.Next()
//
//		duration := time.Since(start).Seconds()
//		status := c.Writer.Status()
//		ctx := c.Request.Context()
//
//		attrs := []attribute.KeyValue{
//			attribute.String("method", c.Request.Method),
//			attribute.String("path", routeLabel(c)),
//			attribute.Int("status", status),
//		}
//
//		requestsCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
//		requestLatency.Record(ctx, duration, metric.WithAttributes(attrs...))
//
//		reqSize := c.Request.ContentLength
//		if reqSize < 0 {
//			reqSize = 0
//		}
//
//		requestSize.Record(ctx, reqSize, metric.WithAttributes(attrs...))
//		responseSize.Record(ctx, int64(c.Writer.Size()), metric.WithAttributes(attrs...))
//	}
//}

func GinZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		req := c.Request
		path := req.URL.Path
		rawQuery := req.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		route := c.FullPath() // 关键：逻辑路由（/users/:id）

		fields := []zap.Field{
			// ---- HTTP 语义（标准字段）----
			zap.String("http.method", req.Method),
			zap.String("http.route", route),
			zap.String("http.target", buildTarget(path, rawQuery)),
			zap.Int("http.status_code", status),
			zap.String("client.ip", c.ClientIP()),
			zap.String("user_agent.original", req.UserAgent()),
			zap.Duration("http.server.duration", latency),
		}

		// ---- 错误信息 ----
		if errs := c.Errors.ByType(gin.ErrorTypePrivate); len(errs) > 0 {
			fields = append(fields,
				zap.String("error.message", errs.String()),
			)
		}

		logger := logging.LoggerFromContext(req.Context())

		// ---- Level 策略（非常重要）----
		switch {
		case status >= 500:
			logger.Error("http request", fields...)
		case status >= 400:
			logger.Warn("http request", fields...)
		case latency > time.Second:
			logger.Warn("slow http request", fields...)
		default:
			logger.Info("http request", fields...)
		}
	}
}

func buildTarget(path, rawQuery string) string {
	if rawQuery == "" {
		return path
	}
	return path + "?" + rawQuery
}

/********************
 * Recovery
 ********************/

func GinZapRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger := logging.LoggerFromContext(c.Request.Context())
				logger.Error("panic recovered",
					zap.Any("panic", rec),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
				)
				c.AbortWithStatusJSON(500, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := logging.InjectLogger(c.Request.Context(), logging.Logger)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
