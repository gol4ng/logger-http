package logger_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/middleware"
)

// Tripperware will decorate the http.Client Transport to add support of gol4ng/logger
func Tripperware(log logger.LoggerInterface, opts ...Option) func(next http.RoundTripper) http.RoundTripper {
	o := evaluateClientOpt(opts...)
	return func(next http.RoundTripper) http.RoundTripper {
		return httpware.RoundTripFunc(func(req *http.Request) (resp *http.Response, err error) {
			startTime := time.Now()

			ctx := req.Context()
			currentLogger := logger.FromContext(ctx, log)
			baseContext := o.loggerContextProvider(req)
			currentContext := feedContext(baseContext, ctx, req, startTime).Add("http_kind", "client")

			if wrappableLogger, ok := currentLogger.(logger.WrappableLoggerInterface); ok {
				currentLogger = wrappableLogger.WrapNew(middleware.Context(currentContext))
			}

			defer func() {
				if err := recover(); err != nil {
					duration := time.Since(startTime)
					currentContext.Add("http_duration", duration.Seconds())
					currentContext.Add("http_panic", err)
					_ = currentLogger.Critical(fmt.Sprintf("http client panic %s %s [duration:%s]", req.Method, req.URL, duration), currentContext)
					panic(err)
				}
			}()

			resp, err = next.RoundTrip(req)
			duration := time.Since(startTime)
			currentContext.Add("http_duration", duration.Seconds())
			if err != nil {
				currentContext.Add("http_error", err)
				currentContext.Add("http_error_message", err.Error())
			}
			if resp == nil {
				_ = currentLogger.Error(fmt.Sprintf("http client error %s %s [duration:%s] %s", req.Method, req.URL, time.Since(startTime), err), currentContext)
				return resp, err
			}
			currentContext.Add("http_status", resp.Status).
				Add("http_status_code", resp.StatusCode).
				Add("http_response_length", resp.ContentLength)

			_ = currentLogger.Log(
				fmt.Sprintf(
					"http client %s %s [status_code:%d, duration:%s, content_length:%d]",
					req.Method,
					req.URL,
					resp.StatusCode,
					duration,
					resp.ContentLength,
				),
				o.levelFunc(resp.StatusCode),
				currentContext,
			)

			return resp, err
		})
	}
}
