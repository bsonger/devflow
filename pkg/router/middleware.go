package router

import (
	"context"
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

/********************
 * Zap Logger
 ********************/

func GinZapLogger(loggerFunc func(ctx context.Context) *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", time.Since(start)),
		}

		if errs := c.Errors.ByType(gin.ErrorTypePrivate); len(errs) > 0 {
			fields = append(fields, zap.String("error", errs.String()))
		}

		loggerFunc(c.Request.Context()).Info("http request", fields...)
	}
}

/********************
 * Recovery
 ********************/

func GinZapRecovery(loggerFunc func(ctx context.Context) *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				loggerFunc(c).Error("panic recovered",
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
