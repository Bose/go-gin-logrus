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
# example log entry for a gin request
{
  "comment": "",
  "fields.time": "2018-07-05T00:08:01Z",
  "ip": "::1",
  "latency": "     13.352µs",
  "level": "info",
  "method": "GET",
  "msg": "",
  "path": "/",
  "status": 200,
  "time": "2018-07-04T20:08:01-04:00",
  "traceIDField": "5035b28a16cd3e8e:5035b28a16cd3e8e:0:1",
  "user-agent": "curl/7.54.0"
}
```

```go
package main

import (
	"fmt"
	"os"
	"time"

	ginlogrus "github.com/Bose/go-gin-logrus"
	"github.com/Bose/go-gin-opentracing"
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
		[]byte("tracing-context"),
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
		time.Sleep(2 * time.Second) // sleep so it's easy to see the timing of entries in the log
		c.JSON(200, "Hello world!")
	})

	r.Run(":29090")
}


```

See the [example.go file](https://github.com/Bose/go-gin-logrus/blob/master/example/example.go)

