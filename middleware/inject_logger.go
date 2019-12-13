package middleware

import (
	"net/http"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/logger"
)

// InjectLogger will inject logger on request context if not exist
// this injection is made for every incoming request so prefer to
// use http.Server in order to create base context
// eg:
//	server := &http.Server{
//		BaseContext: func(listener net.Listener) context.Context {
//			return logger.InjectInContext(context.Background(), l)
//		},
//		...
//	}
func InjectLogger(log logger.LoggerInterface) httpware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			if logger.FromContext(ctx, nil) == nil {
				req = req.WithContext(logger.InjectInContext(ctx, log))
			}
			next.ServeHTTP(writer, req)
		})
	}
}
