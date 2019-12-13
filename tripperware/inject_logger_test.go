package tripperware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gol4ng/logger-http/mocks"
	"github.com/gol4ng/logger-http/tripperware"
)

func TestInjectLogger(t *testing.T) {
	myLogger := logger.NewNopLogger()

	roundTripperMock := &mocks.RoundTripper{}
	req := httptest.NewRequest(http.MethodPost, "http://fake-addr", nil)
	resp := &http.Response{
		Status:        "OK",
		StatusCode:    http.StatusOK,
		ContentLength: 30,
	}

	roundTripperMock.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(resp, nil).Run(func(args mock.Arguments) {
		innerReq := args.Get(0).(*http.Request)
		assert.Equal(t, myLogger, logger.FromContext(innerReq.Context(), nil))
	})

	resp2, err := tripperware.InjectLogger(myLogger)(roundTripperMock).RoundTrip(req)
	assert.Nil(t, err)
	assert.Equal(t, resp, resp2)
}

func TestInjectLogger_AlreadyInjected(t *testing.T) {
	myLogger := logger.NewNopLogger()
	myLogger2 := logger.NewNopLogger()

	roundTripperMock := &mocks.RoundTripper{}
	request := httptest.NewRequest(http.MethodPost, "http://fake-addr", nil)
	response := &http.Response{
		Status:        "OK",
		StatusCode:    http.StatusOK,
		ContentLength: 30,
	}

	roundTripperMock.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(response, nil).Run(func(args mock.Arguments) {
		innerRequest := args.Get(0).(*http.Request)
		// not equal because request.WithContext create another request object
		assert.NotEqual(t, request, innerRequest)
		assert.Equal(t, myLogger, logger.FromContext(innerRequest.Context(), nil))
	})

	stack := httpware.TripperwareStack(
		tripperware.InjectLogger(myLogger),
		tripperware.InjectLogger(myLogger2), // this tripperware not inject logger because logger already injected
	)
	resultResponse, err := stack.DecorateRoundTripper(roundTripperMock).RoundTrip(request)
	assert.Nil(t, err)
	assert.Equal(t, response, resultResponse)
}
