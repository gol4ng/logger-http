package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gol4ng/logger"
	testing_logger "github.com/gol4ng/logger/testing"
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

	myLogger, store := testing_logger.NewLogger()

	middleware.Logger(myLogger)(h).ServeHTTP(responseWriter, request)

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "server", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, int64(2), eCtx["http_response_length"].Value)
		assert.Equal(t, "OK", eCtx["http_status"].Value)
		assert.Equal(t, int64(200), eCtx["http_status_code"].Value)
		assert.Equal(t, "http://127.0.0.1/my-fake-url", eCtx["http_url"].Value)
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}

	entry1 := entries[0]
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "http server received GET http://127.0.0.1/my-fake-url", entry1.Message)

	entry2 := entries[1]
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Regexp(t, `http server GET http://127\.0\.0\.1/my-fake-url \[status_code:200, duration:.*, content_length:.*\]`, entry2.Message)
}

func TestLogger_WithPanic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(req.Context(), 3*time.Second)
	req = req.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("my handler panic")
	})

	myLogger, store := testing_logger.NewLogger()

	assert.PanicsWithValue(t, "my handler panic", func() {
		middleware.Logger(myLogger)(h).ServeHTTP(responseWriter, req)
	})

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "server", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, "http://127.0.0.1/my-fake-url", eCtx["http_url"].Value)
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}

	entry1 := entries[0]
	eCtx1 := *entry1.Context
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "http server received GET http://127.0.0.1/my-fake-url", entry1.Message)
	assert.Equal(t, "my handler panic", eCtx1["http_panic"].Value)

	entry2 := entries[1]
	assert.Equal(t, logger.CriticalLevel, entry2.Level)
	assert.Regexp(t, `http server panic GET http://127\.0\.0\.1/my-fake-url \[duration:.*\]`, entry2.Message)
	assert.Equal(t, "my handler panic", eCtx1["http_panic"].Value)
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

	myLogger, store := testing_logger.NewLogger()

	middleware.Logger(myLogger, logger_http.WithLoggerContext(func(request *http.Request) *logger.Context {
		return logger.NewContext().Add("base_context_key", "base_context_value")
	}))(h).ServeHTTP(responseWriter, request)

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "base_context_value", eCtx["base_context_key"].Value)

		assert.Equal(t, "server", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, int64(2), eCtx["http_response_length"].Value)
		assert.Equal(t, "OK", eCtx["http_status"].Value)
		assert.Equal(t, int64(200), eCtx["http_status_code"].Value)
		assert.Equal(t, "http://127.0.0.1/my-fake-url", eCtx["http_url"].Value)
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}

	entry1 := entries[0]
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "http server received GET http://127.0.0.1/my-fake-url", entry1.Message)

	entry2 := entries[1]
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Regexp(t, `http server GET http://127\.0\.0\.1/my-fake-url \[status_code:200, duration:.*, content_length:.*\]`, entry2.Message)
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

	myLogger, store := testing_logger.NewLogger()

	middleware.Logger(myLogger, logger_http.WithLevels(func(statusCode int) logger.Level {
		return logger.EmergencyLevel
	}))(h).ServeHTTP(responseWriter, request)

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "server", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, int64(2), eCtx["http_response_length"].Value)
		assert.Equal(t, "OK", eCtx["http_status"].Value)
		assert.Equal(t, int64(200), eCtx["http_status_code"].Value)
		assert.Equal(t, "http://127.0.0.1/my-fake-url", eCtx["http_url"].Value)
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}

	entry1 := entries[0]
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "http server received GET http://127.0.0.1/my-fake-url", entry1.Message)

	entry2 := entries[1]
	assert.Equal(t, logger.EmergencyLevel, entry2.Level)
	assert.Regexp(t, `http server GET http://127\.0\.0\.1/my-fake-url \[status_code:200, duration:.*, content_length:.*\]`, entry2.Message)
}
