package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"time"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		duration := time.Since(startTime)
		log.Info().
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Int("status", ctx.Writer.Status()).
			Dur("duration", duration)
	}
}
