package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gol4ng/httpware/v4"
	http_middleware "github.com/gol4ng/httpware/v4/middleware"
	"github.com/gol4ng/logger"

	"github.com/gol4ng/logger-http"
)

// Logger will decorate the http.Handler to add support of gol4ng/logger
func Logger(log logger.LoggerInterface, opts ...logger_http.Option) httpware.Middleware {
	o := logger_http.EvaluateServerOpt(opts...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			startTime := time.Now()
			ctx := req.Context()

			currentLogger := logger.FromContext(ctx, log)
			currentLoggerContext := logger_http.FeedContext(o.LoggerContextProvider(req), ctx, req, startTime).Add("http_kind", "server")

			writerInterceptor := http_middleware.NewResponseWriterInterceptor(writer)
			defer func() {
				duration := time.Since(startTime)
				currentLoggerContext.Add("http_duration", duration.Seconds())

				if err := recover(); err != nil {
					currentLoggerContext.Add("http_panic", err)
					currentLogger.Critical(fmt.Sprintf("http server panic %s %s [duration:%s]", req.Method, req.URL, duration), *currentLoggerContext.Slice()...)
					panic(err)
				}

				currentLoggerContext.Add("http_status", http.StatusText(writerInterceptor.StatusCode)).
					Add("http_status_code", writerInterceptor.StatusCode).
					Add("http_response_length", len(writerInterceptor.Body))

				currentLogger.Log(
					fmt.Sprintf(
						"http server %s %s [status_code:%d, duration:%s, content_length:%d]",
						req.Method, req.URL, writerInterceptor.StatusCode, duration, len(writerInterceptor.Body),
					),
					o.LevelFunc(writerInterceptor.StatusCode),
					*currentLoggerContext.Slice()...,
				)
			}()

			currentLogger.Debug(fmt.Sprintf("http server received %s %s", req.Method, req.URL), *currentLoggerContext.Slice()...)
			next.ServeHTTP(writerInterceptor, req)
		})
	}
}
