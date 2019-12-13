package tripperware

import (
	"fmt"
	"net/http"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/httpware/v2/correlation_id"
	http_tripperware "github.com/gol4ng/httpware/v2/tripperware"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/middleware"

	logger_http "github.com/gol4ng/logger-http"
)

// CorrelationId is a decoration of CorrelationId(github.com/gol4ng/httpware/v2/tripperware)
// it will add correlationId to gol4ng/logger context
// this tripperware require request context with a WrappableLoggerInterface in order to properly add
// correlationID to the logger context
// eg:
//	stack := httpware.TripperwareStack(
//		tripperware.InjectLogger(l), // << Inject logger before CorrelationId
//		tripperware.CorrelationId(),
//	)
// OR
//	ctx := logger.InjectInContext(context.Background(), yourWrappableLogger)
// 	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://a.zz.fr", nil)
//	http.Client{}.Do(req)
func CorrelationId(options ...correlation_id.Option) httpware.Tripperware {
	warning := logger_http.MessageWithFileLine("correlationId need a wrappable logger", 1)
	config := correlation_id.GetConfig(options...)
	orig := http_tripperware.CorrelationId(options...)
	return func(next http.RoundTripper) http.RoundTripper {
		return orig(httpware.RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
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
			return next.RoundTrip(req)
		}))
	}
}
