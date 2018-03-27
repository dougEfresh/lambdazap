# {{ .title.name }}

{{ .title.description }}

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report][report-img]][report]

## Installation 
{{ .installation }}

## Quick Start

{{ .quickStart.code}}

{{ .quickStart.description }}

## Usage 

There are some non context fields such as FunctionName which you can add to all logging requests

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

{{- range .examples }}
    
{{.}}
    
{{- end }}


## Prerequisites

go 1.x

## Tests 

{{- range .tests }}
    
{{.}}
    
{{- end }}


## Benchmarks

In the spirit of Uber's zap logger, zero allocations are used: 

 | Type | Time | Objects Allocated |
 | :--- | :---: | :---: |
 | Non Context | ~150 ns/op | 0 allocs/op
 | With Basic | ~400 ns/op | 0 allocs/op
 | With All | ~733 ns/op | 0 allocs/op

## Deployment

## Contributing
 All PRs are welcome

## Authors

* **Douglas Chimento**  - [{{.user}}][me]

## License

This project is licensed under the Apache License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

* [Uber zap][zap]

### TODO 

[doc-img]: https://godoc.org/github.com/{{.user}}/{{.project}}?status.svg
[doc]: https://godoc.org/github.com/{{.user}}/{{.project}}
[ci-img]: https://travis-ci.org/{{.user}}/{{.project}}.svg?branch=master
[ci]: https://travis-ci.org/{{.user}}/{{.project}}
[cov-img]: https://codecov.io/gh/{{.user}}/{{.project}}/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/{{.user}}/{{.project}}
[glide.lock]: https://github.com/uber-go/zap/blob/master/glide.lock
[zap]: https://github.com/uber-go/zap
[me]: https://github.com/{{.user}}
[report-img]: https://goreportcard.com/badge/github.com/{{.user}}/{{.project}}
[report]: https://goreportcard.com/report/github.com/{{.user}}/{{.project}}