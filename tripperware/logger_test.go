package tripperware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/logger"
	testing_logger "github.com/gol4ng/logger/testing"
	"github.com/stretchr/testify/assert"

	"github.com/gol4ng/logger-http"
	"github.com/gol4ng/logger-http/tripperware"
)

func TestTripperware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/my-fake-url")
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	myLogger, store := testing_logger.NewLogger()

	c := http.Client{
		Transport: tripperware.Logger(myLogger)(http.DefaultTransport),
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/my-fake-url", nil)

	_, err := c.Do(request)
	assert.Nil(t, err)

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "client", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, int64(2), eCtx["http_response_length"].Value)
		assert.Equal(t, "200 OK", eCtx["http_status"].Value)
		assert.Equal(t, int64(200), eCtx["http_status_code"].Value)
		assert.Regexp(t, `http://127\.0\.0\.1.*/my-fake-url`, eCtx["http_url"].Value)
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}
}

func TestTripperware_WithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "server must not be called")
	}))
	defer server.Close()

	myLogger, store := testing_logger.NewLogger()

	c := http.Client{
		Transport: tripperware.Logger(myLogger)(http.DefaultTransport),
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://a.zz/my-fake-url", nil)

	_, err := c.Do(request)
	assert.Contains(t, err.Error(), "no such host")

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "client", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, "dial tcp: lookup a.zz: no such host", eCtx["http_error_message"].Value)
		assert.Contains(t, eCtx, "http_error")
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}

	entry1 := entries[0]
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "http client gonna GET http://a.zz/my-fake-url", entry1.Message)

	entry2 := entries[1]
	assert.Equal(t, logger.ErrorLevel, entry2.Level)
	assert.Regexp(t, `http client error GET http://a\.zz/my-fake-url \[duration:.*\] dial tcp: lookup a\.zz: no such host`, entry2.Message)
}

func TestTripperware_WithPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "server must not be called")
	}))
	defer server.Close()

	myLogger, store := testing_logger.NewLogger()

	transportPanic := httpware.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		panic("my transport panic")
	})
	c := http.Client{
		Transport: tripperware.Logger(myLogger)(transportPanic),
	}

	assert.Panics(t, func() {
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 3*time.Second)
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://a.zz/my-fake-url", nil)

		c.Do(request)
	})

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "my transport panic", eCtx["http_panic"].Value)
		assert.Equal(t, "client", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, "http://a.zz/my-fake-url", eCtx["http_url"].Value)
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}

	entry1 := entries[0]
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "http client gonna GET http://a.zz/my-fake-url", entry1.Message)

	entry2 := entries[1]
	assert.Equal(t, logger.CriticalLevel, entry2.Level)
	assert.Regexp(t, `http client panic GET http://a.zz/my-fake-url \[duration:.*\]`, entry2.Message)
}

func TestTripperware_WithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/my-fake-url")
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	myLogger, store := testing_logger.NewLogger()

	c := http.Client{
		Transport: tripperware.Logger(myLogger, logger_http.WithLoggerContext(func(request *http.Request) *logger.Context {
			return logger.NewContext().Add("base_context_key", "base_context_value")
		}))(http.DefaultTransport),
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/my-fake-url", nil)

	_, err := c.Do(request)
	assert.Nil(t, err)

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "base_context_value", eCtx["base_context_key"].Value)

		assert.Equal(t, "client", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, int64(2), eCtx["http_response_length"].Value)
		assert.Equal(t, "200 OK", eCtx["http_status"].Value)
		assert.Equal(t, int64(200), eCtx["http_status_code"].Value)
		assert.Regexp(t, `http://127\.0\.0\.1.*/my-fake-url`, eCtx["http_url"].Value)
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}
}

func TestTripperware_WithLevels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/my-fake-url")
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	myLogger, store := testing_logger.NewLogger()

	c := http.Client{
		Transport: tripperware.Logger(myLogger, logger_http.WithLevels(func(statusCode int) logger.Level {
			return logger.EmergencyLevel
		}))(http.DefaultTransport),
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/my-fake-url", nil)

	_, err := c.Do(request)
	assert.Nil(t, err)

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "client", eCtx["http_kind"].Value)
		assert.Equal(t, "GET", eCtx["http_method"].Value)
		assert.Equal(t, int64(2), eCtx["http_response_length"].Value)
		assert.Equal(t, "200 OK", eCtx["http_status"].Value)
		assert.Equal(t, int64(200), eCtx["http_status_code"].Value)
		assert.Regexp(t, `http://127\.0\.0\.1.*/my-fake-url`, eCtx["http_url"].Value)
		assert.Contains(t, eCtx, "http_duration")
		assert.Contains(t, eCtx, "http_start_time")
		assert.Contains(t, eCtx, "http_request_deadline")
	}

	entry1 := entries[0]
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Regexp(t, `http client gonna GET http://127\.0\.0\.1.*/my-fake-url`, entry1.Message)

	entry2 := entries[1]
	assert.Equal(t, logger.EmergencyLevel, entry2.Level)
	assert.Regexp(t, `http client GET http://127\.0\.0\.1.*/my-fake-url \[status_code:200, duration:.*, content_length:.*\]`, entry2.Message)
}
