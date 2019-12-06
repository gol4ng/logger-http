package tripperware

import (
	"net/http"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/httpware/v2/correlation_id"
	tripperware2 "github.com/gol4ng/httpware/v2/tripperware"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/middleware"
)

// CorrelationId will decorate the http.Handler to add support of gol4ng/logger
func CorrelationId(log logger.LoggerInterface, options ...correlation_id.Option) httpware.Tripperware {
	config := correlation_id.GetConfig(options...)
	orig := tripperware2.CorrelationId(options...)
	return func(next http.RoundTripper) http.RoundTripper {
		return orig(httpware.RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			ctx := req.Context()
			currentLogger := logger.FromContext(ctx, log)
			if wrappableLogger, ok := currentLogger.(logger.WrappableLoggerInterface); ok {
				logger.InjectInContext(ctx, wrappableLogger.Wrap(
					middleware.Context(
						logger.NewContext().Add(config.HeaderName, ctx.Value(config.HeaderName)),
					),
				))
			}
			return next.RoundTrip(req)
		}))
	}
}
