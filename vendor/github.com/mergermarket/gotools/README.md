# gotools

Gotools provides a few simple building blocks for Mergermarket apps.

## internal

Internal endpoints for http services. 

Example usage:

```
	router := http.NewServeMux()
	router.HandleFunc("/internal/healthcheck", tools.InternalHealthCheck)
	router.HandleFunc("/internal/log-config", tools.NewInternalLogConfig(config, log))

```

## logger

Logger logs messages in a structured format in AWS and pretty colours in local.

Example usage:

```
  logger := tools.NewLogger(config.IsLocal())
	logger.Info("Hello!")

```

## statsd

StatsD provides an interface for all of the DataDog StatsD methods used by mergermarket. 

Example usage:

```
	statsdConfig := tools.NewStatsDConfig(!config.IsLocal(), logger)
	statsd, err := tools.NewStatsD(statsdConfig)
	if err != nil {
		logger.Error("Error connecting to StatsD. Stats will only be logged. Error: ", err.Error())
	}
  statsd.Histogram("important_action", .0001, "tag1:tag1value", "tag2:tag2value")

```

## telemetry-handler

WrapWithTelemetry takes your http.Handler and adds debug logs with request details and marks metrics

Example usage:

```
  router := http.NewServeMux()
  router.Handle("/my-important-endpoint", importantHandler)  
  tools.WrapWithTelemetry("/", router, logger, statsd)
```

## test tools

```
testLogger, testStatsd := tools.NewTestTools(t)
```
