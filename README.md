# go-gin-logrus
[![](https://godoc.org/github.com/Bose/go-gin-logrus?status.svg)](https://godoc.org/github.com/Bose/go-gin-logrus)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bose/go-gin-logrus)](https://goreportcard.com/report/github.com/Bose/go-gin-logrus)
[![Release](https://img.shields.io/github/release/Bose/go-gin-logrus.svg?style=flat-square)](https://Bose/go-gin-logrus/releases)

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
# example aggregated log entry for a request with UseBanner == true
{
  "new-header-index-name": "this is how you set new header level data",
  "request-summary-info": {
    "comment": "",
    "ip": "::1",
    "latency": "     98.217µs",
    "method": "GET",
    "path": "/",
    "requestID": "4b4fb22ef51cc540:4b4fb22ef51cc540:0:1",
    "status": 200,
    "time": "2019-02-06T13:24:06Z",
    "user-agent": "curl/7.54.0"
  },
  "entries": [
    {
      "level": "info",
      "msg": "this will be aggregated into one write with the access log and will show up when the request is completed",
      "time": "2019-02-06T08:24:06-05:00"
    },
    {
      "comment": "this is an aggregated log entry with initial comment field",
      "level": "debug",
      "msg": "aggregated entry with new comment field",
      "time": "2019-02-06T08:24:06-05:00"
    },
    {
      "level": "error",
      "msg": "aggregated error entry with new-comment field",
      "new-comment": "this is an aggregated log entry with reset comment field",
      "time": "2019-02-06T08:24:06-05:00"
    }
  ],
  "banner": "[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------"
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

	tracer, reporter, closer, err := ginopentracing.InitTracing(
		fmt.Sprintf("go-gin-logrus-example::%s", hostName), // service name for the traces
		"localhost:5775",                        // where to send the spans
		ginopentracing.WithEnableInfoLog(false)) // WithEnableLogInfo(false) will not log info on every span sent... if set to true it will log and they won't be aggregated
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
		ginlogrus.SetCtxLoggerHeader(c, "new-header-index-name", "this is how you set new header level data")

		logger := ginlogrus.GetCtxLogger(c) // will get a logger with the aggregate Logger set if it's enabled - handy if you've already set fields for the request
		logger.Info("this will be aggregated into one write with the access log and will show up when the request is completed")

		// add some new fields to the existing logger
		logger = ginlogrus.SetCtxLogger(c, logger.WithFields(logrus.Fields{"comment": "this is an aggregated log entry with initial comment field"}))
		logger.Debug("aggregated entry with new comment field")

		// replace existing logger fields with new ones (notice it's logrus.WithFields())
		logger = ginlogrus.SetCtxLogger(c, logrus.WithFields(logrus.Fields{"new-comment": "this is an aggregated log entry with reset comment field"}))
		logger.Error("aggregated error entry with new-comment field")

		logrus.Info("this will NOT be aggregated and will be logged immediately")
		span := newSpanFromContext(c, "sleep-span")
		defer span.Finish() // this will get logged because tracing was setup with ginopentracing.WithEnableInfoLog(true)

		go func() {
			// need a NewBuffer for aggregate logging of this goroutine (since the req will be done long before this thing finishes)
			// it will inherit header info from the existing request
			buff := ginlogrus.NewBuffer(logger)
			time.Sleep(1 * time.Second)
			logger.Info("Hi from a goroutine completing after the request")
			fmt.Printf(buff.String())
		}()
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

