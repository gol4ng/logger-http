package middleware

import (
	"net/http"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/httpware/v2/correlation_id"
	middleware2 "github.com/gol4ng/httpware/v2/middleware"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/middleware"
)

// CorrelationId will decorate the http.Handler to add support of gol4ng/logger
func CorrelationId(log logger.LoggerInterface, options ...correlation_id.Option) httpware.Middleware {
	config := correlation_id.GetConfig(options...)
	orig := middleware2.CorrelationId(options...)
	return func(next http.Handler) http.Handler {
		return orig(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			currentLogger := logger.FromContext(ctx, log)
			if wrappableLogger, ok := currentLogger.(logger.WrappableLoggerInterface); ok {
				logger.InjectInContext(ctx, wrappableLogger.Wrap(
					middleware.Context(
						logger.NewContext().Add(config.HeaderName, ctx.Value(config.HeaderName)),
					),
				))
			}
			next.ServeHTTP(writer, req)
		}))
	}
}
