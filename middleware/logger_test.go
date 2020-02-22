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

	entry1 := entries[0]
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `http server received GET http://127.0.0.1/my-fake-url`, entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http server GET http://127.0.0.1/my-fake-url [status_code:200, duration:`)
	assert.Contains(t, entry2.Message, `content_length:2]`)
	assert.Equal(t, "http://127.0.0.1/my-fake-url", (*entry2.Context)["http_url"].Value)
	assert.Equal(t, int64(2), (*entry2.Context)["http_response_length"].Value)
	assert.Equal(t, "OK", (*entry2.Context)["http_status"].Value)
	assert.Equal(t, int64(200), (*entry2.Context)["http_status_code"].Value)
	assert.Contains(t, *entry2.Context, "http_duration")
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

	entry1 := entries[0]
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `http server received GET http://127.0.0.1/my-fake-url`, entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.CriticalLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http server panic GET http://127.0.0.1/my-fake-url [duration:`)
	assert.Equal(t, "GET", (*entry2.Context)["http_method"].Value)
	assert.Equal(t, "http://127.0.0.1/my-fake-url", (*entry2.Context)["http_url"].Value)

	assert.Contains(t, *entry2.Context, `http_panic`)
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

	entry1 := entries[0]
	assert.Equal(t, "base_context_value", (*entry1.Context)["base_context_key"].Value)
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `http server received GET http://127.0.0.1/my-fake-url`, entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	assert.Equal(t, "base_context_value", (*entry2.Context)["base_context_key"].Value)
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http server GET http://127.0.0.1/my-fake-url [status_code:200, duration:`)
	assert.Contains(t, entry2.Message, `content_length:2]`)
	assert.Equal(t, "http://127.0.0.1/my-fake-url", (*entry2.Context)["http_url"].Value)
	assert.Equal(t, int64(2), (*entry2.Context)["http_response_length"].Value)
	assert.Equal(t, "OK", (*entry2.Context)["http_status"].Value)
	assert.Equal(t, int64(200), (*entry2.Context)["http_status_code"].Value)
	assert.Contains(t, *entry2.Context, "http_duration")
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

	entry1 := entries[0]
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `http server received GET http://127.0.0.1/my-fake-url`, entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.EmergencyLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http server GET http://127.0.0.1/my-fake-url [status_code:200, duration:`)
	assert.Contains(t, entry2.Message, `content_length:2]`)
	assert.Equal(t, "http://127.0.0.1/my-fake-url", (*entry2.Context)["http_url"].Value)
	assert.Equal(t, int64(2), (*entry2.Context)["http_response_length"].Value)
	assert.Equal(t, "OK", (*entry2.Context)["http_status"].Value)
	assert.Equal(t, int64(200), (*entry2.Context)["http_status_code"].Value)
	assert.Contains(t, *entry2.Context, "http_duration")
}

func AssertDefaultContextFields(t *testing.T, entry logger.Entry) {
	assert.Equal(t, "server", (*entry.Context)["http_kind"].Value)
	assert.Contains(t, *entry.Context, "http_method")
	assert.Contains(t, *entry.Context, "http_url")
	assert.Contains(t, *entry.Context, "http_start_time")
	assert.Contains(t, *entry.Context, "http_request_deadline")
}
