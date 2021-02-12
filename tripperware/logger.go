package tripperware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gol4ng/httpware/v4"
	"github.com/gol4ng/logger"

	"github.com/gol4ng/logger-http"
)

// Logger will decorate the http.Client to add support of gol4ng/logger
func Logger(log logger.LoggerInterface, opts ...logger_http.Option) func(next http.RoundTripper) http.RoundTripper {
	o := logger_http.EvaluateClientOpt(opts...)
	return func(next http.RoundTripper) http.RoundTripper {
		return httpware.RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			startTime := time.Now()
			ctx := req.Context()

			currentLogger := logger.FromContext(ctx, log)
			currentLoggerContext := logger_http.FeedContext(o.LoggerContextProvider(req), ctx, req, startTime).Add("http_kind", "client")

			defer func() {
				duration := time.Since(startTime)
				currentLoggerContext.Add("http_duration", duration.Seconds())

				if err := recover(); err != nil {
					currentLoggerContext.Add("http_panic", err)
					currentLogger.Critical(fmt.Sprintf("http client panic %s %s [duration:%s]", req.Method, req.URL, duration), *currentLoggerContext.Slice()...)
					panic(err)
				}
				if err != nil {
					currentLoggerContext.Add("http_error", err)
					currentLoggerContext.Add("http_error_message", err.Error())
				}
				if resp == nil {
					currentLogger.Error(fmt.Sprintf("http client error %s %s [duration:%s] %s", req.Method, req.URL, duration, err), *currentLoggerContext.Slice()...)
					return
				}
				currentLoggerContext.Add("http_status", resp.Status).
					Add("http_status_code", resp.StatusCode).
					Add("http_response_length", resp.ContentLength)

				currentLogger.Log(
					fmt.Sprintf(
						"http client %s %s [status_code:%d, duration:%s, content_length:%d]",
						req.Method, req.URL, resp.StatusCode, duration, resp.ContentLength,
					),
					o.LevelFunc(resp.StatusCode),
					*currentLoggerContext.Slice()...,
				)
			}()

			currentLogger.Debug(fmt.Sprintf("http client gonna %s %s", req.Method, req.URL), *currentLoggerContext.Slice()...)
			return next.RoundTrip(req)
		})
	}
}
