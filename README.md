# logger-http

[![Build Status](https://travis-ci.org/gol4ng/logger-http.svg?branch=master)](https://travis-ci.org/gol4ng/logger-http)
[![Go Report Card](https://goreportcard.com/badge/github.com/gol4ng/logger-http)](https://goreportcard.com/report/github.com/gol4ng/logger-http)
[![GoDoc](https://godoc.org/github.com/gol4ng/logger-http?status.svg)](https://godoc.org/github.com/gol4ng/logger-http)

Gol4ng logger sub package for logging http
Related package [gol4ng/logger](https://github.com/gol4ng/logger) and [gol4ng/httpware](https://github.com/gol4ng/httpware)  

## Installation

`go get -u github.com/gol4ng/logger-http`

## Quick Start

You can refer at [gol4ng/httpware](https://github.com/gol4ng/httpware) documentation for the middleware/tripperware usage

### Tripperware 

Log you're `http.Client` request

```
<debug> http client gonna GET http://google.com {"http_url":"http://google.com","http_start_time":"2019-12-13T17:01:13+01:00","http_kind":"client","Correlation-Id":"yhyBI94zyl","http_header":{"Correlation-Id":["yhyBI94zyl"]},"http_method":"GET"}
<info> http client GET http://google.com [status_code:301, duration:39.70726ms, content_length:219] {"Correlation-Id":"yhyBI94zyl","http_start_time":"2019-12-13T17:01:13+01:00","http_method":"GET","http_duration":0.03970726,"http_response_length":219,"http_header":{"Correlation-Id":["yhyBI94zyl"]},"http_url":"http://google.com","http_status":"301 Moved Permanently","http_status_code":301,"http_kind":"client"}
<debug> http client gonna GET http://www.google.com/ {"http_kind":"client","Correlation-Id":"uQzaMO9JC0","http_header":{"Correlation-Id":["uQzaMO9JC0"],"Referer":["http://google.com"]},"http_method":"GET","http_url":"http://www.google.com/","http_start_time":"2019-12-13T17:01:13+01:00"}
<info> http client GET http://www.google.com/ [status_code:200, duration:72.582736ms, content_length:-1] {"Correlation-Id":"uQzaMO9JC0","http_header":{"Correlation-Id":["uQzaMO9JC0"],"Referer":["http://google.com"]},"http_method":"GET","http_kind":"client","http_response_length":-1,"http_duration":0.072582736,"http_status_code":200,"http_url":"http://www.google.com/","http_start_time":"2019-12-13T17:01:13+01:00","http_status":"200 OK"}
```

```go
package main

import (
	"net/http"
	"os"

	"github.com/gol4ng/httpware/v2"
	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger-http/tripperware"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
)

func main(){
	// logger will print on STDOUT with default line format
	myLogger := logger.NewLogger(handler.Stream(os.Stdout, formatter.NewDefaultFormatter()))

	clientStack := httpware.TripperwareStack(
		tripperware.InjectLogger(myLogger),
		tripperware.CorrelationId(),
		tripperware.Logger(myLogger),
	)

	c := http.Client{
		Transport: clientStack.DecorateRoundTripper(http.DefaultTransport),
	}

	c.Get("http://google.com")
	// Will log
	//<info> http client GET http://google.com [status_code:301, duration:27.524999ms, content_length:219] {"http_duration":0.027524999,"http_status":"301 Moved Permanently","http_status_code":301,"http_response_length":219,"http_method":"GET","http_url":"http://google.com","http_start_time":"2019-12-03T10:47:38+01:00","http_kind":"client"}
	//<info> http client GET http://www.google.com/ [status_code:200, duration:51.047002ms, content_length:-1] {"http_kind":"client","http_duration":0.051047002,"http_status":"200 OK","http_status_code":200,"http_response_length":-1,"http_method":"GET","http_url":"http://www.google.com/","http_start_time":"2019-12-03T10:47:38+01:00"}
}
```

### Middleware
Log you're incoming http server request

```
<debug> http server received GET / {"http_url":"/","http_start_time":"2019-12-13T17:15:30+01:00","http_kind":"server","Correlation-Id":"SBeEdhRhUl","http_header":{"Accept-Encoding":["gzip"],"Correlation-Id":["SBeEdhRhUl"],"User-Agent":["Go-http-client/1.1"]},"http_method":"GET"}
<info> handler log info {"Correlation-Id":"SBeEdhRhUl"}
<info> http server GET / [status_code:200, duration:290.156µs, content_length:0] {"http_kind":"server","http_duration":0.000290156,"http_status_code":200,"Correlation-Id":"SBeEdhRhUl","http_method":"GET","http_url":"/","http_start_time":"2019-12-13T17:15:30+01:00","http_status":"OK","http_response_length":0,"http_header":{"Accept-Encoding":["gzip"],"Correlation-Id":["SBeEdhRhUl"],"User-Agent":["Go-http-client/1.1"]}}
```

```go
package main
     
     import (
     	"context"
     	"net"
     	"net/http"
     	"os"
     
     	"github.com/gol4ng/httpware/v2"
     	"github.com/gol4ng/logger"
     	"github.com/gol4ng/logger-http/middleware"
     	"github.com/gol4ng/logger/formatter"
     	"github.com/gol4ng/logger/handler"
     )
     
     func main() {
     	addr := ":5001"
     
     	myLogger := logger.NewLogger(
     		handler.Stream(os.Stdout, formatter.NewDefaultFormatter()),
     	)
     
     	// we recommend to use MiddlewareStack to simplify managing all wanted middlewares
     	// caution middleware order matters
     	stack := httpware.MiddlewareStack(
     		//middleware.InjectLogger(myLogger), // we recommend to use http.Server.BaseContext instead of this middleware
     		middleware.CorrelationId(),
     		middleware.Logger(myLogger),
     	)
     
     	h := http.HandlerFunc(func(writer http.ResponseWriter, innerRequest *http.Request) {
     		l := logger.FromContext(innerRequest.Context(), myLogger)
     		l.Info("handler log info", nil)
     	})
     
     	server := http.Server{
     		Addr:    addr,
     		Handler: stack.DecorateHandler(h),
     		BaseContext: func(listener net.Listener) context.Context {
     			return logger.InjectInContext(context.Background(), myLogger)
     		},
     	}
     
     	go func() {
     		if err := server.ListenAndServe(); err != nil {
     			panic(err)
     		}
     	}()
     
     	http.Get("http://localhost" + addr)
     
     	//<debug> http server received GET / {"Correlation-Id":"zEDWO9gmZ6","http_header":{"Accept-Encoding":["gzip"],"Correlation-Id":["zEDWO9gmZ6"],"User-Agent":["Go-http-client/1.1"]},"http_method":"GET","http_url":"/","http_start_time":"2019-12-13T17:05:53+01:00","http_kind":"server"}
     	//<info> handler log info {"Correlation-Id":"zEDWO9gmZ6"}
     	//<info> http server GET / [status_code:200, duration:232.491µs, content_length:0] {"Correlation-Id":"zEDWO9gmZ6","http_header":{"Accept-Encoding":["gzip"],"Correlation-Id":["zEDWO9gmZ6"],"User-Agent":["Go-http-client/1.1"]},"http_method":"GET","http_url":"/","http_start_time":"2019-12-13T17:05:53+01:00","http_duration":0.000232491,"http_status":"OK","http_status_code":200,"http_kind":"server","http_response_length":0}
     }
```
