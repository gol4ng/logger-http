package logger_http

import (
	"context"
	"net/http"
	"time"

	"github.com/gol4ng/logger"
)

type Options struct {
	LoggerContextProvider LoggerContextProvider
	LevelFunc             CodeToLevel
}

// LoggerContextProvider function defines the default logger context values
type LoggerContextProvider func(*http.Request) *logger.Context

// CodeToLevel function defines the mapping between http.StatusCode and logger.Level
type CodeToLevel func(statusCode int) logger.Level

var (
	defaultOptions = &Options{
		LoggerContextProvider: func(request *http.Request) *logger.Context {
			return logger.NewContext().
				Add("http_header", request.Header).
				Add("http_body", request.GetBody)
		},
		LevelFunc: func(statusCode int) logger.Level {
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

func EvaluateClientOpt(opts ...Option) *Options {
	optCopy := &Options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func EvaluateServerOpt(opts ...Option) *Options {
	optCopy := &Options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*Options)

// WithLoggerContext will provide default logger context values
func WithLoggerContext(f LoggerContextProvider) Option {
	return func(o *Options) {
		o.LoggerContextProvider = f
	}
}

// WithLevels customizes the function for the mapping between http.StatusCode and logger.Level
func WithLevels(f CodeToLevel) Option {
	return func(o *Options) {
		o.LevelFunc = f
	}
}

func FeedContext(loggerContext *logger.Context, ctx context.Context, req *http.Request, startTime time.Time) *logger.Context {
	if loggerContext == nil {
		loggerContext = logger.NewContext()
	}
	loggerContext.
		Add("http_method", req.Method).
		Add("http_url", req.URL.String()).
		Add("http_start_time", startTime.Format(time.RFC3339))

	if d, ok := ctx.Deadline(); ok {
		loggerContext.Add("http_request_deadline", d.Format(time.RFC3339))
	}
	return loggerContext
}
