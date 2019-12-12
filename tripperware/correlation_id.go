package tripperware

import (
	"net/http"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/httpware/v2/correlation_id"
	http_tripperware "github.com/gol4ng/httpware/v2/tripperware"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/middleware"
)

// CorrelationId will decorate the http.Handler to add support of gol4ng/logger
func CorrelationId(log logger.WrappableLoggerInterface, options ...correlation_id.Option) httpware.Tripperware {
	config := correlation_id.GetConfig(options...)
	orig := http_tripperware.CorrelationId(options...)
	return func(next http.RoundTripper) http.RoundTripper {
		return orig(httpware.RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
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

			return next.RoundTrip(req)
		}))
	}
}
