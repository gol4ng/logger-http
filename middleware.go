package logger_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gol4ng/httpware/v2"
	middleware2 "github.com/gol4ng/httpware/v2/middleware"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/middleware"
)

// Middleware will decorate the http.Handler to add support of gol4ng/logger
func Middleware(log logger.LoggerInterface, opts ...Option) httpware.Middleware {
	o := evaluateClientOpt(opts...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			startTime := time.Now()

			ctx := req.Context()
			currentLogger := logger.FromContext(ctx, log)
			baseContext := o.loggerContextProvider(req)
			currentContext := feedContext(baseContext, ctx, req, startTime).Add("http_kind", "server")

			if wrappableLogger, ok := currentLogger.(logger.WrappableLoggerInterface); ok {
				currentLogger = wrappableLogger.WrapNew(middleware.Context(currentContext))
			}
			writerInterceptor := middleware2.NewResponseWriterInterceptor(writer)
			defer func() {
				duration := time.Since(startTime)
				currentContext.Add("http_duration", duration.Seconds())

				if err := recover(); err != nil {
					currentContext.Add("http_panic", err)
					_ = currentLogger.Critical(fmt.Sprintf("http server panic %s %s [duration:%s]", req.Method, req.URL, duration), currentContext)
					panic(err)
				}

				currentContext.Add("http_status", http.StatusText(writerInterceptor.StatusCode)).
					Add("http_status_code", writerInterceptor.StatusCode).
					Add("http_response_length", len(writerInterceptor.Body))

				_ = currentLogger.Log(
					fmt.Sprintf(
						"http server %s %s [status_code:%d, duration:%s, content_length:%d]",
						req.Method,
						req.URL,
						writerInterceptor.StatusCode,
						duration,
						len(writerInterceptor.Body),
					),
					o.levelFunc(writerInterceptor.StatusCode),
					currentContext,
				)
			}()

			next.ServeHTTP(writerInterceptor, req)
		})
	}
}
