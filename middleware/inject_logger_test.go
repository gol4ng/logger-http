package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gol4ng/httpware/v4"
	"github.com/gol4ng/logger"
	"github.com/stretchr/testify/assert"

	"github.com/gol4ng/logger-http/middleware"
)

func TestInjectLogger(t *testing.T) {
	myLogger := logger.NewNopLogger()

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(request.Context(), 3*time.Second)
	request = request.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	h := http.HandlerFunc(func(writer http.ResponseWriter, innerRequest *http.Request) {
		requestLogger := logger.FromContext(innerRequest.Context(), nil)
		// not equal because request.WithContext create another request object
		assert.NotEqual(t, request, innerRequest)
		assert.Equal(t, myLogger, requestLogger)
		writer.Write([]byte(`OK`))
	})

	middleware.InjectLogger(myLogger)(h).ServeHTTP(responseWriter, request)
}

func TestInjectLogger_AlreadyInjected(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/my-fake-url", nil)
	ctx, _ := context.WithTimeout(request.Context(), 3*time.Second)
	request = request.WithContext(ctx)
	responseWriter := &httptest.ResponseRecorder{}

	myLogger := logger.NewNopLogger()
	myLogger2 := logger.NewNopLogger()

	h := func(writer http.ResponseWriter, innerRequest *http.Request) {
		// not equal because request.WithContext create another request object
		assert.NotEqual(t, request, innerRequest)
		assert.Equal(t, myLogger, logger.FromContext(innerRequest.Context(), nil))
		writer.Write([]byte(`OK`))
	}

	httpware.MiddlewareStack(
		middleware.InjectLogger(myLogger),
		middleware.InjectLogger(myLogger2), // this tripperware not inject logger because logger already injected
	).DecorateHandlerFunc(h).ServeHTTP(responseWriter, request)
}
