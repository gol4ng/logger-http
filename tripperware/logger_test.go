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

	entry1 := entries[0]
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `http client gonna GET `+server.URL+`/my-fake-url`, entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http client GET `+server.URL+`/my-fake-url [status_code:200, duration:`)
	assert.Contains(t, entry2.Message, `content_length:2]`)
	assert.Equal(t, server.URL+"/my-fake-url", (*entry2.Context)["http_url"].Value)
	assert.Equal(t, int64(2), (*entry2.Context)["http_response_length"].Value)
	assert.Equal(t, "200 OK", (*entry2.Context)["http_status"].Value)
	assert.Equal(t, int64(200), (*entry2.Context)["http_status_code"].Value)
	assert.Contains(t, *entry2.Context, "http_duration")
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

	entry1 := entries[0]
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "http client gonna GET http://a.zz/my-fake-url", entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.ErrorLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http client error GET http://a.zz/my-fake-url [duration:`)
	assert.Contains(t, entry2.Message, `dial tcp: lookup a.zz`)
	assert.Equal(t, "http://a.zz/my-fake-url", (*entry2.Context)["http_url"].Value)
	assert.Contains(t, *entry2.Context, "http_duration")
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

	entry1 := entries[0]
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "http client gonna GET http://a.zz/my-fake-url", entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.CriticalLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http client panic GET http://a.zz/my-fake-url [duration:`)
	assert.Equal(t, "http://a.zz/my-fake-url", (*entry2.Context)["http_url"].Value)
	assert.Contains(t, *entry2.Context, "http_duration")
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

	entry1 := entries[0]
	assert.Equal(t, "base_context_value", (*entry1.Context)["base_context_key"].Value)
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `http client gonna GET `+server.URL+`/my-fake-url`, entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	assert.Equal(t, "base_context_value", (*entry2.Context)["base_context_key"].Value)
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http client GET `+server.URL+`/my-fake-url [status_code:200, duration:`)
	assert.Contains(t, entry2.Message, `content_length:2]`)
	assert.Equal(t, server.URL+"/my-fake-url", (*entry2.Context)["http_url"].Value)
	assert.Equal(t, int64(2), (*entry2.Context)["http_response_length"].Value)
	assert.Equal(t, "200 OK", (*entry2.Context)["http_status"].Value)
	assert.Equal(t, int64(200), (*entry2.Context)["http_status_code"].Value)
	assert.Contains(t, *entry2.Context, "http_duration")
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

	entry1 := entries[0]
	AssertDefaultContextFields(t, entry1)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `http client gonna GET `+server.URL+`/my-fake-url`, entry1.Message)
	assert.Equal(t, "GET", (*entry1.Context)["http_method"].Value)

	entry2 := entries[1]
	AssertDefaultContextFields(t, entry2)
	assert.Equal(t, logger.EmergencyLevel, entry2.Level)
	assert.Contains(t, entry2.Message, `http client GET `+server.URL+`/my-fake-url [status_code:200, duration:`)
	assert.Contains(t, entry2.Message, `content_length:2]`)
	assert.Equal(t, server.URL+"/my-fake-url", (*entry2.Context)["http_url"].Value)
	assert.Equal(t, int64(2), (*entry2.Context)["http_response_length"].Value)
	assert.Equal(t, "200 OK", (*entry2.Context)["http_status"].Value)
	assert.Equal(t, int64(200), (*entry2.Context)["http_status_code"].Value)
	assert.Contains(t, *entry2.Context, "http_duration")
}

func AssertDefaultContextFields(t *testing.T, entry logger.Entry) {
	assert.Equal(t, "client", (*entry.Context)["http_kind"].Value)
	assert.Contains(t, *entry.Context, "http_method")
	assert.Contains(t, *entry.Context, "http_url")
	assert.Contains(t, *entry.Context, "http_start_time")
	assert.Contains(t, *entry.Context, "http_request_deadline")
}
