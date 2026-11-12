// Package httpserver contains the shared Gin setup and graceful HTTP server runner.
package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthPath is served by every HTTP service for liveness checks.
const HealthPath = "/healthz"

// NewRouter creates a Gin engine with structured request logging, panic recovery
// and a health endpoint.
func NewRouter(log *slog.Logger, production bool) *gin.Engine {
	if production {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(RequestLogger(log), Recovery(log))
	router.GET(HealthPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return router
}

// RequestLogger logs one line per request with its status and latency.
func RequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if c.Request.URL.Path == HealthPath {
			return
		}

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= http.StatusInternalServerError:
			level = slog.LevelError
		case status >= http.StatusBadRequest:
			level = slog.LevelWarn
		}

		log.LogAttrs(c.Request.Context(), level, "http request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
		)
	}
}

// Recovery turns panics into a logged 500 response instead of crashing the service.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, err any) {
		log.ErrorContext(c.Request.Context(), "panic recovered", slog.Any("error", err))
		Error(c, http.StatusInternalServerError, "Internal server error")
	})
}

// Error aborts the request with a JSON body of the form {"message": "..."}.
func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"message": message})
}
