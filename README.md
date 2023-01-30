# Admin Console

## Overview

## Dependencies

## Setup for local development
1. This project uses **go modules** so to install all dependencies you need to run
`go mod download && go mod vendor`
2. Configure environment in default.env
3. Compile and run

## Project structure
- *app* - applications which contains main business logic
- *controller* - k8s controllers https://kubernetes.io/docs/concepts/architecture/controller/
- *templates* - html templates for html rendering https://pkg.go.dev/html/template
- *config* - app config structure
- *service* - external services clients

## Errors & Logging
The project uses pkg/errors lib to handle errors and stack traces. Every error must be wrapped as follows:
```
if err != nil {
  return nil, errors.Wrap(err, "unable to get registry list")
}
```
Wraping an error added context to the underlying error and recorded the file and 
line that the error occurred. This file and line information could be retrieved via a 
helper function, Fprint, to give a trace of the execution path leading away from the error.

**zap** - Blazing fast, structured, leveled logging in Go.
To use logging you need to get logger instance from `service/logger` package:
```
var log = logger.GetLogger()
```
There is different logging level functions provided by logger interface:
- **Info** - Info logs a message at InfoLevel;
- **Warn** - Warn logs a message at WarnLevel;
- **Error** - Error logs a message at ErrorLevel;
- **Debug** - Debug logs a message at DebugLevel;
- **DPanic** - DPanic logs a message at DPanicLevel. If the logger is in development mode, it then panics 
(DPanic means "development panic"). This is useful for catching errors that are recoverable, but shouldn't ever happen;
- **Panic** - Panic logs a message at PanicLevel. The logger then panics, even if logging at PanicLevel is disabled;
- **Fatal** - Fatal logs a message at FatalLevel; The logger then calls os.Exit(1), 
even if logging at FatalLevel is disabled.

## HTTP Handling & Routing

## Go tests
Calculate test coverage
```
go test -coverprofile=coverage.out && go tool cover -html=coverage.out && rm -rf coverage.out
```


## Port forward for local development

- `oc port-forward -n control-plane service/gerrit 31000:31000 8090:8080`
- `oc port-forward -n control-plane service/jenkins 8099:8080`
- `oc port-forward -n user-management service/hashicorp-vault 8200:8200 8201:8201`