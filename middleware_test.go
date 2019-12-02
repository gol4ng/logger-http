package logger_http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger-http"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(req.Context(), 3*time.Second)
	req = req.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// not equal because req.WithContext create another request object
		assert.NotEqual(t, responseWriter, w)
		w.Write([]byte(`OK`))
	})

	output := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(output, formatter.NewDefaultFormatter()),
	)

	logger_http.Middleware(myLogger)(h).ServeHTTP(responseWriter, req)
	output.Constains(t, []string{
		`<info> http server GET http://127.0.0.1/my-fake-url [status_code:200, duration:`,
		`content_length:2] {`,
		`"http_kind":"server"`,
		`"http_method":"GET"`,
		`"http_response_length":2`,
		`"http_status":"OK"`,
		`"http_status_code":200`,

		`"http_url":"http://127.0.0.1`,
		`"http_duration":`,
		`"http_start_time":"`,
		`"http_request_deadline":"`,
	})
}

func TestMiddleware_WithPanic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(req.Context(), 3*time.Second)
	req = req.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("my handler panic")
	})

	output := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(output, formatter.NewDefaultFormatter()),
	)

	assert.PanicsWithValue(t, "my handler panic", func() {
		logger_http.Middleware(myLogger)(h).ServeHTTP(responseWriter, req)
	})

	output.Constains(t, []string{
		`<critical> http server panic GET http://127.0.0.1/my-fake-url [duration:`,
		`"http_kind":"server"`,
		`"http_method":"GET"`,

		`"http_url":"http://127.0.0.1`,
		`"http_duration":`,
		`"http_start_time":"`,
		`"http_request_deadline":"`,
	})
}

func TestMiddleware_WithContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(req.Context(), 3*time.Second)
	req = req.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// not equal because req.WithContext create another request object
		assert.NotEqual(t, responseWriter, w)
		w.Write([]byte(`OK`))
	})

	output := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(output, formatter.NewDefaultFormatter()),
	)

	logger_http.Middleware(myLogger, logger_http.WithLoggerContext(func(request *http.Request) *logger.Context {
		return logger.NewContext().Add("base_context_key", "base_context_value")
	}))(h).ServeHTTP(responseWriter, req)

	output.Constains(t, []string{
		`<info> http server GET http://127.0.0.1/my-fake-url [status_code:200, duration:`,
		`content_length:2] {`,
		`"http_kind":"server"`,
		`"http_method":"GET"`,
		`"http_response_length":2`,
		`"http_status":"OK"`,
		`"http_status_code":200`,
		`"base_context_key":"base_context_value"`,

		`"http_url":"http://127.0.0.1`,
		`"http_duration":`,
		`"http_start_time":"`,
		`"http_request_deadline":"`,
	})
}

func TestMiddleware_WithLevels(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(req.Context(), 3*time.Second)
	req = req.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// not equal because req.WithContext create another request object
		assert.NotEqual(t, responseWriter, w)
		w.Write([]byte(`OK`))
	})

	output := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(output, formatter.NewDefaultFormatter()),
	)

	logger_http.Middleware(myLogger, logger_http.WithLevels(func(statusCode int) logger.Level {
		return logger.EmergencyLevel
	}))(h).ServeHTTP(responseWriter, req)

	output.Constains(t, []string{
		`<emergency> http server GET http://127.0.0.1/my-fake-url [status_code:200, duration:`,
		`content_length:2] {`,
		`"http_kind":"server"`,
		`"http_method":"GET"`,
		`"http_response_length":2`,
		`"http_status":"OK"`,
		`"http_status_code":200`,

		`"http_url":"http://127.0.0.1`,
		`"http_duration":`,
		`"http_start_time":"`,
		`"http_request_deadline":"`,
	})
}
