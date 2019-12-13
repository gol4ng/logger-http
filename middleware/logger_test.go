package middleware_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
	"github.com/stretchr/testify/assert"

	"github.com/gol4ng/logger-http"
	"github.com/gol4ng/logger-http/middleware"
)

func TestLogger(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(request.Context(), 3*time.Second)
	request = request.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(writer http.ResponseWriter, innerRequest *http.Request) {
		// not equal because request.WithContext create another request object
		assert.NotEqual(t, responseWriter, writer)
		writer.Write([]byte(`OK`))
	})

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

	middleware.Logger(myLogger)(h).ServeHTTP(responseWriter, request)
	loggerOutput.Constains(t, []string{
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

func TestLogger_WithPanic(t *testing.T) {
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
		middleware.Logger(myLogger)(h).ServeHTTP(responseWriter, req)
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

func TestLogger_WithContext(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(request.Context(), 3*time.Second)
	request = request.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(writer http.ResponseWriter, innerRequest *http.Request) {
		// not equal because request.WithContext create another request object
		assert.NotEqual(t, responseWriter, writer)
		writer.Write([]byte(`OK`))
	})

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

	middleware.Logger(myLogger, logger_http.WithLoggerContext(func(request *http.Request) *logger.Context {
		return logger.NewContext().Add("base_context_key", "base_context_value")
	}))(h).ServeHTTP(responseWriter, request)

	loggerOutput.Constains(t, []string{
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

func TestLogger_WithLevels(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(request.Context(), 3*time.Second)
	request = request.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(writer http.ResponseWriter, innerRequest *http.Request) {
		// not equal because request.WithContext create another request object
		assert.NotEqual(t, responseWriter, writer)
		writer.Write([]byte(`OK`))
	})

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

	middleware.Logger(myLogger, logger_http.WithLevels(func(statusCode int) logger.Level {
		return logger.EmergencyLevel
	}))(h).ServeHTTP(responseWriter, request)

	loggerOutput.Constains(t, []string{
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

type Output struct {
	bytes.Buffer
}

func (o *Output) Constains(t *testing.T, str []string) {
	b := o.String()
	for _, s := range str {
		if strings.Contains(b, s) != true {
			assert.Fail(t, fmt.Sprintf("buffer %s must contain %s\n", b, s))
		}
	}
}
