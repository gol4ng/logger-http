package tripperware

import (
	"net/http"

	"github.com/gol4ng/httpware/v3"
	"github.com/gol4ng/logger"
)

// InjectLogger will inject logger on request context if not exist
func InjectLogger(log logger.LoggerInterface) httpware.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return httpware.RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			ctx := req.Context()
			if logger.FromContext(ctx, nil) == nil {
				req = req.WithContext(logger.InjectInContext(ctx, log))
			}
			return next.RoundTrip(req)
		})
	}
}
