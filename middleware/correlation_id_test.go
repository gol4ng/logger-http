package middleware_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/httpware/v2/correlation_id"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
	"github.com/stretchr/testify/assert"

	http_middleware "github.com/gol4ng/logger-http/middleware"
)

func TestCorrelationId(t *testing.T) {
	correlation_id.DefaultIdGenerator = correlation_id.NewRandomIdGenerator(
		rand.New(correlation_id.NewLockedSource(rand.NewSource(1))),
	)

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

	request := httptest.NewRequest(http.MethodGet, "http://fake-addr", nil)
	request = request.WithContext(logger.InjectInContext(request.Context(), myLogger))
	responseWriter := &httptest.ResponseRecorder{}

	var handlerRequest *http.Request
	h := http.HandlerFunc(func(writer http.ResponseWriter, innerRequest *http.Request) {
		// not equal because request.WithContext create another request object
		assert.NotEqual(t, request, innerRequest)
		assert.Equal(t, "p1LGIehp1s", innerRequest.Header.Get(correlation_id.HeaderName))
		handlerRequest = innerRequest
		logger.FromContext(innerRequest.Context(), nil).Info("handler info log", nil)
	})

	myLogger.Info("info log before request", logger.NewContext().Add("ctxvalue", "before"))
	http_middleware.CorrelationId()(h).ServeHTTP(responseWriter, request)
	myLogger.Info("info log after request", logger.NewContext().Add("ctxvalue", "after"))

	respHeaderValue := responseWriter.Header().Get(correlation_id.HeaderName)
	reqContextValue := handlerRequest.Context().Value(correlation_id.HeaderName).(string)
	assert.Equal(t, "p1LGIehp1s", request.Header.Get(correlation_id.HeaderName))
	assert.True(t, len(respHeaderValue) == 10)
	assert.True(t, len(reqContextValue) == 10)
	assert.True(t, respHeaderValue == reqContextValue)
	loggerOutput.Constains(t, []string{
		`<info> info log before request {"ctxvalue":"before"}`,
		`<info> handler info log {"Correlation-Id":"p1LGIehp1s"}`,
		`<info> info log after request {"ctxvalue":"after"}`,
	})
}

func TestCorrelationId_WithoutWrappableLogger(t *testing.T) {
	correlation_id.DefaultIdGenerator = correlation_id.NewRandomIdGenerator(
		rand.New(correlation_id.NewLockedSource(rand.NewSource(1))),
	)

	loggerOutput := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(loggerOutput, formatter.NewDefaultFormatter()),
	)

	request := httptest.NewRequest(http.MethodGet, "http://fake-addr", nil)
	// WE DO NOT INJECT LOGGER IN REQUEST CONTEXT
	//request = request.WithContext(logger.InjectInContext(request.Context(), myLogger))
	responseRecorder := &httptest.ResponseRecorder{}

	var handlerRequest *http.Request
	h := http.HandlerFunc(func(writer http.ResponseWriter, innerRequest *http.Request) {
		// not equal because request.WithContext create another request object
		assert.NotEqual(t, request, innerRequest)
		assert.Equal(t, "p1LGIehp1s", innerRequest.Header.Get(correlation_id.HeaderName))
		handlerRequest = innerRequest
		assert.Nil(t, logger.FromContext(innerRequest.Context(), nil))
	})

	myLogger.Info("info log before request", logger.NewContext().Add("ctxvalue", "before"))
	output := getStdout(func() {
		http_middleware.CorrelationId()(h).ServeHTTP(responseRecorder, request)
	})
	myLogger.Info("info log after request", logger.NewContext().Add("ctxvalue", "after"))

	respHeaderValue := responseRecorder.Header().Get(correlation_id.HeaderName)
	reqContextValue := handlerRequest.Context().Value(correlation_id.HeaderName).(string)
	assert.Equal(t, "p1LGIehp1s", request.Header.Get(correlation_id.HeaderName))
	assert.True(t, len(respHeaderValue) == 10)
	assert.True(t, len(reqContextValue) == 10)
	assert.True(t, respHeaderValue == reqContextValue)
	assert.Contains(t, output, "correlationId need a wrappable logger /")
	assert.Contains(t, output, "/src/github.com/gol4ng/logger-http/middleware/correlation_id_test.go:")

	loggerOutput.Constains(t, []string{
		`<info> info log before request {"ctxvalue":"before"}`,
		`<info> info log after request {"ctxvalue":"after"}`,
	})
}

// Use to get os.Stdout
var mu = sync.Mutex{}

func lock() func() {
	mu.Lock()
	return func() {
		mu.Unlock()
	}
}

func getStdout(f func()) string {
	defer lock()()
	// keep original os.Stdout
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	// replace os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = orig
	}()

	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

// =====================================================================================================================
// ========================================= EXAMPLES ==================================================================
// =====================================================================================================================

func ExampleCorrelationId() {
	port := ":5001"

	myLogger := logger.NewLogger(
		handler.Stream(os.Stdout, formatter.NewDefaultFormatter()),
	)

	// we recommend to use MiddlewareStack to simplify managing all wanted middlewares
	// caution middleware order matters
	stack := httpware.MiddlewareStack(
		http_middleware.InjectLogger(myLogger),
		http_middleware.CorrelationId(
			correlation_id.WithHeaderName("my-personal-header-name"),
			correlation_id.WithIdGenerator(func(request *http.Request) string {
				return "my-fixed-request-id"
			}),
		),
	)

	h := http.HandlerFunc(func(writer http.ResponseWriter, innerRequest *http.Request) {
		l := logger.FromContext(innerRequest.Context(), myLogger)
		l.Info("handler log info", nil)
	})

	go func() {
		if err := http.ListenAndServe(port, stack.DecorateHandler(h)); err != nil {
			panic(err)
		}
	}()

	resp, err := http.Get("http://localhost" + port)
	fmt.Printf("%s: %v %v\n", "my-personal-header-name", resp.Header.Get("my-personal-header-name"), err)

	//Output:
	// <info> handler log info {"my-personal-header-name":"my-fixed-request-id"}
	// my-personal-header-name: my-fixed-request-id <nil>
}
