package tripperware_test

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gol4ng/httpware/v4"
	"github.com/gol4ng/httpware/v4/correlation_id"
	"github.com/gol4ng/logger"
	testing_logger "github.com/gol4ng/logger/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gol4ng/logger-http/mocks"
	"github.com/gol4ng/logger-http/tripperware"
)

func TestCorrelationId(t *testing.T) {
	correlation_id.DefaultIdGenerator = correlation_id.NewRandomIdGenerator(
		rand.New(correlation_id.NewLockedSource(rand.NewSource(1))),
	)

	myLogger, store := testing_logger.NewLogger()

	roundTripperMock := &mocks.RoundTripper{}
	request := httptest.NewRequest(http.MethodPost, "http://fake-addr", nil)
	request = request.WithContext(logger.InjectInContext(request.Context(), myLogger))
	response := &http.Response{
		Status:        "OK",
		StatusCode:    http.StatusOK,
		ContentLength: 30,
	}

	var handlerReq *http.Request
	roundTripperMock.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(response, nil).Run(func(args mock.Arguments) {
		innerRequest := args.Get(0).(*http.Request)
		assert.NotEqual(t, request, innerRequest)
		assert.Equal(t, "p1LGIehp1s", innerRequest.Header.Get(correlation_id.HeaderName))
		handlerReq = innerRequest
		logger.FromContext(innerRequest.Context(), nil).Info("handler info log")
	})

	myLogger.Info("info log before request", logger.String("ctxvalue", "before"))
	resultResponse, err := tripperware.CorrelationId()(roundTripperMock).RoundTrip(request)
	myLogger.Info("info log after request", logger.String("ctxvalue", "after"))

	assert.Nil(t, err)
	assert.Equal(t, response, resultResponse)
	assert.Equal(t, "p1LGIehp1s", handlerReq.Header.Get(correlation_id.HeaderName))
	assert.Equal(t, "", response.Header.Get(correlation_id.HeaderName))

	entries := store.GetEntries()
	assert.Len(t, entries, 3)

	entry1 := entries[0]
	assert.Equal(t, logger.InfoLevel, entry1.Level)
	assert.Equal(t, "info log before request", entry1.Message)
	assert.Equal(t, "before", (*entry1.Context)["ctxvalue"].Value)

	entry2 := entries[1]
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Equal(t, "handler info log", entry2.Message)
	assert.Equal(t, "p1LGIehp1s", (*entry2.Context)["Correlation-Id"].Value)

	entry3 := entries[2]
	assert.Equal(t, logger.InfoLevel, entry3.Level)
	assert.Equal(t, "info log after request", entry3.Message)
	assert.Equal(t, "after", (*entry3.Context)["ctxvalue"].Value)
}

func TestCorrelationId_WithoutWrappableLogger(t *testing.T) {
	correlation_id.DefaultIdGenerator = correlation_id.NewRandomIdGenerator(
		rand.New(correlation_id.NewLockedSource(rand.NewSource(1))),
	)

	myLogger, store := testing_logger.NewLogger()

	roundTripperMock := &mocks.RoundTripper{}
	request := httptest.NewRequest(http.MethodPost, "http://fake-addr", nil)
	// WE DO NOT INJECT LOGGER IN REQUEST CONTEXT
	//request = request.WithContext(logger.InjectInContext(request.Context(), myLogger))
	response := &http.Response{
		Status:        "OK",
		StatusCode:    http.StatusOK,
		ContentLength: 30,
	}

	var handlerRequest *http.Request
	roundTripperMock.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(response, nil).Run(func(args mock.Arguments) {
		innerRequest := args.Get(0).(*http.Request)
		assert.NotEqual(t, request, innerRequest)
		assert.Equal(t, "p1LGIehp1s", innerRequest.Header.Get(correlation_id.HeaderName))
		handlerRequest = innerRequest
		assert.Nil(t, logger.FromContext(innerRequest.Context(), nil))
	})

	var resultResponse *http.Response
	var err error
	myLogger.Info("info log before request", logger.String("ctxvalue", "before"))
	output := getStdout(func() {
		resultResponse, err = tripperware.CorrelationId()(roundTripperMock).RoundTrip(request)
	})
	myLogger.Info("info log after request", logger.String("ctxvalue", "after"))

	assert.Nil(t, err)
	assert.Equal(t, response, resultResponse)
	assert.Equal(t, "p1LGIehp1s", handlerRequest.Header.Get(correlation_id.HeaderName))
	assert.Equal(t, "", response.Header.Get(correlation_id.HeaderName))
	assert.Contains(t, output, "correlationId need a wrappable logger /")
	assert.Contains(t, output, "/src/github.com/gol4ng/logger-http/tripperware/correlation_id_test.go")

	entries := store.GetEntries()
	assert.Len(t, entries, 2)

	entry1 := entries[0]
	assert.Equal(t, logger.InfoLevel, entry1.Level)
	assert.Equal(t, "info log before request", entry1.Message)
	assert.Equal(t, "before", (*entry1.Context)["ctxvalue"].Value)

	entry2 := entries[1]
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Equal(t, "info log after request", entry2.Message)
	assert.Equal(t, "after", (*entry2.Context)["ctxvalue"].Value)
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
	myLogger := logger.NewNopLogger()

	clientStack := httpware.TripperwareStack(
		tripperware.InjectLogger(myLogger),
		tripperware.CorrelationId(
			correlation_id.WithHeaderName("my-personal-header-name"),
			correlation_id.WithIdGenerator(func(request *http.Request) string {
				return "my-fixed-request-id"
			}),
		),
	)

	c := http.Client{
		Transport: clientStack.DecorateRoundTripper(http.DefaultTransport),
	}

	c.Get("http://google.com")
	// Output:
}
