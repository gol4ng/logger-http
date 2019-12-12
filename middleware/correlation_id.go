package middleware

import (
	"net/http"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/httpware/v2/correlation_id"
	http_middleware "github.com/gol4ng/httpware/v2/middleware"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/middleware"
)

// CorrelationId will decorate the http.Handler to add support of gol4ng/logger
func CorrelationId(log logger.WrappableLoggerInterface, options ...correlation_id.Option) httpware.Middleware {
	config := correlation_id.GetConfig(options...)
	orig := http_middleware.CorrelationId(options...)
	return func(next http.Handler) http.Handler {
		return orig(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			loggerContextMiddleware := middleware.Context(
				logger.NewContext().Add(config.HeaderName, ctx.Value(config.HeaderName)),
			)
			var decoratedLogger logger.LoggerInterface
			requestLogger := logger.FromContext(ctx, nil)
			if requestLogger != nil {
				if wrappableLogger, ok := requestLogger.(logger.WrappableLoggerInterface); ok {
					decoratedLogger = wrappableLogger.Wrap(loggerContextMiddleware)
				} else {
					_ = log.Notice("logger not wrappable correlationId not added to request logger", nil)
				}
			} else {
				decoratedLogger = log.WrapNew(loggerContextMiddleware)
			}
			logger.InjectInContext(ctx, decoratedLogger)
			next.ServeHTTP(writer, req)
		}))
	}
}
