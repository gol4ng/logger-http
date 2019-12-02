package logger_http

import (
	"context"
	"net/http"
	"time"

	"github.com/gol4ng/logger"
)

type options struct {
	loggerContextProvider LoggerContextProvider
	levelFunc             CodeToLevel
}

// LoggerContextProvider function defines the default logger context values
type LoggerContextProvider func(*http.Request) *logger.Context

// CodeToLevel function defines the mapping between http.StatusCode and logger.Level
type CodeToLevel func(statusCode int) logger.Level

var (
	defaultOptions = &options{
		loggerContextProvider: func(_ *http.Request) *logger.Context {
			return nil
		},
		levelFunc: func(statusCode int) logger.Level {
			switch {
			case statusCode < http.StatusBadRequest:
				return logger.InfoLevel
			case statusCode < http.StatusInternalServerError:
				return logger.WarningLevel
			}
			return logger.ErrorLevel
		},
	}
)

func evaluateClientOpt(opts ...Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// WithLoggerContext will provide default logger context values
func WithLoggerContext(f LoggerContextProvider) Option {
	return func(o *options) {
		o.loggerContextProvider = f
	}
}

// WithLevels customizes the function for the mapping between http.StatusCode and logger.Level
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

func feedContext(baseContext *logger.Context, ctx context.Context, req *http.Request, startTime time.Time) *logger.Context {
	c := logger.NewContext().
		Add("http_method", req.Method).
		Add("http_url", req.URL.String()).
		Add("http_start_time", startTime.Format(time.RFC3339))

	if baseContext != nil {
		c.Merge(*baseContext)
	}
	if d, ok := ctx.Deadline(); ok {
		c.Add("http_request_deadline", d.Format(time.RFC3339))
	}
	return c
}
