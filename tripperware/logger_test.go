package tripperware_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"

	"github.com/gol4ng/logger-http"
	"github.com/gol4ng/logger-http/tripperware"
)

func TestTripperware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/my-fake-url")
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

	c := http.Client{
		Transport: tripperware.Logger(myLogger)(http.DefaultTransport),
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/my-fake-url", nil)

	_, err := c.Do(request)
	assert.Nil(t, err)
	loggerOutput.Constains(t, []string{
		`<info> http client GET http://127.0.0.1`,
		`/my-fake-url [status_code:200, duration:`,
		`content_length:2] {`,
		`"http_kind":"client"`,
		`"http_method":"GET"`,
		`"http_response_length":2`,
		`"http_status":"200 OK"`,
		`"http_status_code":200`,

		`"http_url":"http://127.0.0.1`,
		`"http_duration":`,
		`"http_start_time":"`,
		`"http_request_deadline":"`,
	})
}

func TestTripperware_WithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "server must not be called")
	}))
	defer server.Close()

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

	c := http.Client{
		Transport: tripperware.Logger(myLogger)(http.DefaultTransport),
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://a.zz/my-fake-url", nil)

	_, err := c.Do(request)
	assert.Contains(t, err.Error(), "no such host")
	loggerOutput.Constains(t, []string{
		`<error> http client error GET http://a.zz/my-fake-url [duration:`,
		`dial tcp: lookup a.zz`,
		`no such host`,
		`"http_kind":"client"`,
		`"http_method":"GET"`,
		`"http_url":"http://a.zz/my-fake-url"`,

		`"http_error_message":"dial tcp: lookup a.zz`,
		`"http_error":`,

		`"http_start_time":"`,
		`"http_request_deadline":"`,
		`"http_duration":`,
	})
}

func TestTripperware_WithPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "server must not be called")
	}))
	defer server.Close()

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

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

	loggerOutput.Constains(t, []string{
		`<critical> http client panic GET http://a.zz/my-fake-url [duration:`,
		`"http_kind":"client"`,
		`"http_method":"GET"`,
		`"http_url":"http://a.zz/my-fake-url"`,
		`"http_panic":"my transport panic"`,

		`"http_start_time":"`,
		`"http_request_deadline":"`,
		`"http_duration":`,
	})
}

func TestTripperware_WithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/my-fake-url")
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

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
	loggerOutput.Constains(t, []string{
		`<info> http client GET http://127.0.0.1`,
		`/my-fake-url [status_code:200, duration:`,
		`content_length:2] {`,
		`"http_kind":"client"`,
		`"http_method":"GET"`,
		`"http_response_length":2`,
		`"http_status":"200 OK"`,
		`"http_status_code":200`,

		`"http_url":"http://127.0.0.1`,
		`"http_duration":`,
		`"http_start_time":"`,
		`"http_request_deadline":"`,
		`"base_context_key":"base_context_value"`,
	})
}

func TestTripperware_WithLevels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/my-fake-url")
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

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
	loggerOutput.Constains(t, []string{
		`<emergency> http client GET http://127.0.0.1`,
		`/my-fake-url [status_code:200, duration:`,
		`content_length:2] {`,
		`"http_kind":"client"`,
		`"http_method":"GET"`,
		`"http_response_length":2`,
		`"http_status":"200 OK"`,
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
