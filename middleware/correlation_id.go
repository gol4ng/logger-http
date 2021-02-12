package middleware

import (
	"fmt"
	"net/http"

	"github.com/gol4ng/httpware/v4"
	"github.com/gol4ng/httpware/v4/correlation_id"
	http_middleware "github.com/gol4ng/httpware/v4/middleware"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/middleware"

	logger_http "github.com/gol4ng/logger-http"
)

// CorrelationId is a decoration of CorrelationId(github.com/gol4ng/httpware/v2/middleware)
// it will add correlationId to gol4ng/logger context
// this middleware require request context with a WrappableLoggerInterface in order to properly add
// correlationID to the logger context
// eg:
//	stack := httpware.MiddlewareStack(
//		middleware.InjectLogger(l), // << Inject logger before CorrelationId
//		middleware.CorrelationId(),
//	)
func CorrelationId(options ...correlation_id.Option) httpware.Middleware {
	warning := logger_http.MessageWithFileLine("correlationId need a wrappable logger", 1)
	config := correlation_id.GetConfig(options...)
	orig := http_middleware.CorrelationId(options...)
	return func(next http.Handler) http.Handler {
		return orig(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			requestLogger := logger.FromContext(ctx, nil)
			injected := false
			if requestLogger != nil {
				if wrappableLogger, ok := requestLogger.(logger.WrappableLoggerInterface); ok {
					req = req.WithContext(logger.InjectInContext(ctx, wrappableLogger.WrapNew(middleware.Context(
						logger.NewContext().Add(config.HeaderName, ctx.Value(config.HeaderName)),
					))))
					injected = true
				}
			}
			if !injected {
				fmt.Println(warning)
			}
			next.ServeHTTP(writer, req)
		}))
	}
}
