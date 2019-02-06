# go-gin-logrus
[![](https://godoc.org/github.com/Bose/go-gin-logrus?status.svg)](https://godoc.org/github.com/Bose/go-gin-logrus) 

Gin Web Framework Open Tracing middleware.

This middleware also support aggregate logging: the ability to aggregate all log entries into just one write.  This aggregation is helpful when your logs are being sent to Kibana and you only want to index one log per request.

## Installation

`$ go get github.com/Bose/go-gin-logrus`

If you want to use it with opentracing you could consider installing:

`$ go get github.com/Bose/go-gin-opentracing`

## Dependencies - for local development
If you want to see your traces on your local system, you'll need to run a tracing backend like Jaeger.   You'll find info about how-to in the [Jaeger Tracing github repo docs](https://github.com/jaegertracing/documentation/blob/master/content/docs/getting-started.md)
Basically, you can run the Jaeger opentracing backend under docker via:

```bash 
docker run -d -e \
  COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 9411:9411 \
  jaegertracing/all-in-one:latest
  ```
## Usage
```
# example aggregated log entry for a request
{
  "entries": [
    {
      "level": "info",
      "method": "GET",
      "msg": "this will be aggregated into one write with the access log and will show up when the request is completed",
      "path": "/",
      "requestID": "7a94e88fce04a760:7a94e88fce04a760:0:1",
      "time": "2019-02-05T21:01:05-05:00"
    },
    {
      "comment": "this is an aggregated log entry",
      "level": "debug",
      "method": "GET",
      "msg": "aggregated entry with new comment field",
      "path": "/",
      "requestID": "7a94e88fce04a760:7a94e88fce04a760:0:1",
      "time": "2019-02-05T21:01:05-05:00"
    },
    {
      "level": "error",
      "method": "GET",
      "msg": "aggregated error entry with new-comment field",
      "new-comment": "this is an aggregated log entry",
      "path": "/",
      "requestID": "7a94e88fce04a760:7a94e88fce04a760:0:1",
      "time": "2019-02-05T21:01:05-05:00"
    },
    {
      "comment": "",
      "fields.time": "2019-02-06T02:01:07Z",
      "ip": "::1",
      "latency": " 2.002710403s",
      "level": "info",
      "method": "GET",
      "msg": "[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------",
      "path": "/",
      "requestID": "7a94e88fce04a760:7a94e88fce04a760:0:1",
      "status": 200,
      "time": "2019-02-05T21:01:07-05:00",
      "user-agent": "curl/7.54.0"
    }
  ]
}

```

```go
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/opentracing/opentracing-go/ext"

	ginlogrus "github.com/Bose/go-gin-logrus"
	ginopentracing "github.com/Bose/go-gin-opentracing"
	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
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

	tracer, reporter, closer, err := ginopentracing.InitTracing(fmt.Sprintf("go-gin-logrus-example::%s", hostName), "localhost:5775", ginopentracing.WithEnableInfoLog(true))
	if err != nil {
		panic("unable to init tracing")
	}
	defer closer.Close()
	defer reporter.Close()
	opentracing.SetGlobalTracer(tracer)

	p := ginopentracing.OpenTracer([]byte("api-request-"))
	r.Use(p)

	r.Use(gin.Recovery()) // add Recovery middleware
	useBanner := true
	useUTC := true
	r.Use(ginlogrus.WithTracing(logrus.StandardLogger(),
		useBanner,
		time.RFC3339,
		useUTC,
		"requestID",
		[]byte("uber-trace-id"), // where jaeger might have put the trace id
		[]byte("RequestID"),     // where the trace ID might already be populated in the headers
		ginlogrus.WithAggregateLogging(true)))

	r.GET("/", func(c *gin.Context) {
		logger := ginlogrus.GetCtxLogger(c) // will get a logger with the aggregate Logger set if it's enabled - handy if you've already set fields for the request
		logger.Info("this will be aggregated into one write with the access log and will show up when the request is completed")

		// add some new fields to the existing logger
		logger = ginlogrus.SetCtxLogger(c, logger.WithFields(logrus.Fields{"comment": "this is an aggregated log entry"}))
		logger.Debug("aggregated entry with new comment field")

		// replace existing logger fields with new ones (notice it's logrus.WithFields())
		logger = ginlogrus.SetCtxLogger(c, logrus.WithFields(logrus.Fields{"new-comment": "this is an aggregated log entry"}))
		logger.Error("aggregated error entry with new-comment field")

		logrus.Info("this will NOT be aggregated and will be logged immediately")
		span := newSpanFromContext(c, "sleep-span")
		defer span.Finish()
		time.Sleep(2 * time.Second) // sleep so it's easy to see the timing of entries in the log
		c.JSON(200, "Hello world!")
	})

	r.Run(":29090")
}

func newSpanFromContext(c *gin.Context, operationName string) opentracing.Span {
	parentSpan, _ := c.Get("tracing-context")
	options := []opentracing.StartSpanOption{
		opentracing.Tag{Key: ext.SpanKindRPCServer.Key, Value: ext.SpanKindRPCServer.Value},
		opentracing.Tag{Key: string(ext.HTTPMethod), Value: c.Request.Method},
		opentracing.Tag{Key: string(ext.HTTPUrl), Value: c.Request.URL.Path},
		opentracing.Tag{Key: "current-goroutines", Value: runtime.NumGoroutine()},
	}

	if parentSpan != nil {
		options = append(options, opentracing.ChildOf(parentSpan.(opentracing.Span).Context()))
	}

	return opentracing.StartSpan(operationName, options...)
}

```

See the [example.go file](https://github.com/Bose/go-gin-logrus/blob/master/example/example.go)

