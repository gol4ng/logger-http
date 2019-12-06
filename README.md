# logger-http

[![Build Status](https://travis-ci.org/gol4ng/logger-http.svg?branch=master)](https://travis-ci.org/gol4ng/logger-http)
[![GoDoc](https://godoc.org/github.com/gol4ng/logger-http?status.svg)](https://godoc.org/github.com/gol4ng/logger-http)

Gol4ng logger sub package for http

## Installation

`go get -u github.com/gol4ng/logger-http`

## Quick Start

Log you `http.Client` request

```go
package main

import (
	"net/http"
	"os"

	"github.com/gol4ng/logger"
	"github.com/gol4ng/logger-http"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
)

func main(){
	// logger will print on STDOUT with default line format
	myLogger := logger.NewLogger(handler.Stream(os.Stdout, formatter.NewDefaultFormatter()))

	c := http.Client{
		Transport: logger_http.Tripperware(myLogger)(http.DefaultTransport),
	}

	c.Get("http://google.com")
    // Will log
    //<info> http client GET http://google.com [status_code:301, duration:27.524999ms, content_length:219] {"http_duration":0.027524999,"http_status":"301 Moved Permanently","http_status_code":301,"http_response_length":219,"http_method":"GET","http_url":"http://google.com","http_start_time":"2019-12-03T10:47:38+01:00","http_kind":"client"}
    //<info> http client GET http://www.google.com/ [status_code:200, duration:51.047002ms, content_length:-1] {"http_kind":"client","http_duration":0.051047002,"http_status":"200 OK","http_status_code":200,"http_response_length":-1,"http_method":"GET","http_url":"http://www.google.com/","http_start_time":"2019-12-03T10:47:38+01:00"}
}
```

Log you incoming http server request
