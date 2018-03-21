# Lambda Zap

Add [AWS lamba context](https://github.com/aws/aws-lambda-go) fields to [uber's zap](https://github.com/uber-go/zap)

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report][report-img]][report]

## Installation 
```shell
$ go get -u github.com/dougEfresh/lambdazap
```

## Quick Start

```go

package main
import (
    "context"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/dougEfresh/lambdazap"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Create a new lambda log context and use RequestID, FunctionName, InvokeFunctionArn and a variable from environment
var lambdazapper = lambdazap.New().
    With(lambdazap.AwsRequestID, lambdazap.FunctionName, lambdazap.InvokeFunctionArn).
    WithEnv("ZAP_TEST")

var logger *zap.Logger

func init() {
    // Init the logger outside of the handler
    logger, _ := zap.NewProduction()
}

func Handler(ctx context.Context) (string, error) {
     defer logger.Sync()
    logger.Info("Starting hander with context values ", lambdazapper.ContextValues(ctx)...)
    return "Uber zap with lambda context", nil
}

func main() {
    lambda.Start(Handler)
}

```




## Usage 

There are some non-context fields such as FunctionName which you can add to all logging requests

```go
lambdazapper := New(lambdazap.ProcessNonContextFields(false)).With(lambdazap.FunctionName, lambdazap.FunctionVersion, lambdazap.AwsRequestID)
logger.With(lambdazapper.NonContextValues()...))
logger.Info("only non context values")
```

The above will log FunctionName and FunctionVersion but *not* RequestId. 

The Non Context fields are 
```shell
	FunctionName
	FunctionVersion
	LogGroupName
	LogStreamName
	MemoryLimitInMB

```

*Note* by default all context and non context will be logged. 
The option `lambdazap.ProcessNonContextFields(false)` will NOT log non context values (e.g. FunctionName) when used like this
```go
logger.Info("only context values. No FunctionName!", lambdazapper.ContextValues()...)
```

## Examples
    
See example [handler](test/handler.go) with [cloudformation](test/test-template.yaml).
    
[List of fields](https://godoc.org/github.com/dougEfresh/lambdazap#LambdaField)


## Prerequisites

go 1.x

## Tests
    
```shell
$ go test -v

```

    
See [travis.yaml](.travis.yml) for running benchmark tests

See [travis.yaml](.travis.yml) for running benchmark tests

## Benchmarks

In the spirit of Uber's zap logger, zero allocations are used: 

 | Type | Time | Objects Allocated |
 | :--- | :---: | :---: |
 | Non Context | ~150 ns/op | 0 allocs/op
 | WithBasic | ~400 ns/op | 0 allocs/op
 | WithAll | ~733 ns/op | 0 allocs/op

## Deployment

## Contributing
 All PRs are welcome

## Authors

* **Douglas Chimento**  - [dougEfresh][me]

## License

This project is licensed under the Apache License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

* [Uber zap][zap]

### TODO 

[doc-img]: https://godoc.org/github.com/dougEfresh/lambdazap?status.svg
[doc]: https://godoc.org/github.com/dougEfresh/lambdazap
[ci-img]: https://travis-ci.org/dougEfresh/lambdazap.svg?branch=master
[ci]: https://travis-ci.org/dougEfresh/lambdazap
[cov-img]: https://codecov.io/gh/dougEfresh/lambdazap/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/dougEfresh/lambdazap
[glide.lock]: https://github.com/uber-go/zap/blob/master/glide.lock
[zap]: https://github.com/uber-go/zap
[me]: https://github.com/dougEfresh
[report-img]: https://goreportcard.com/badge/github.com/dougEfresh/lambdazap
[report]: https://goreportcard.com/report/github.com/dougEfresh/lambdazap