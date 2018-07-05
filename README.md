# go-gin-prometheus
[![](https://godoc.org/github.com/Bose/go-gin-logrus?status.svg)](https://godoc.org/github.com/Bose/go-gin-logrus) 

Gin Web Framework Open Tracing middleware

## Installation

`$ go get github.com/Bose/go-gin-logrus`

## Usage

```go
package main

import (
	"fmt"
	"os"
	"time"

	ginlogrus "github.com/Bose/go-gin-logrus"
	"github.com/sirupsen/logrus"

	"github.com/Bose/go-gin-opentracing"
	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
)

func main() {
	// use the JSON formatter
	logrus.SetFormatter(&logrus.JSONFormatter{})

	r := gin.New() // don't use the Default(), since it comes with a logger

	// setup tracing...
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}

	tracer, closer, err := ginopentracing.Config.New(fmt.Sprintf("go-gin-logrus-example.go::%s", hostName))
	if err == nil {
		fmt.Println("Setting global tracer")
		defer closer.Close()
		opentracing.SetGlobalTracer(tracer)
	} else {
		fmt.Println("Can't enable tracing: ", err.Error())
	}
	p := ginopentracing.OpenTracer([]byte("api-request-"))
	r.Use(p)

	r.Use(gin.Recovery()) // add Recovery middleware
	useBanner := true
	useUTC := true
	r.Use(ginlogrus.WithTracing(logrus.StandardLogger(),
		useBanner,
		time.RFC3339,
		useUTC,
		"traceIDField",
		[]byte("uber-trace-id"),
		[]byte("tracing-context")))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, "Hello world!")
	})

	r.Run(":29090")
}


```

See the [example.go file](https://github.com/github.com/Bose/go-gin-logrus/blob/master/example/example.go)

