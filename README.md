# go-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/zcong1993/kit.svg)](https://pkg.go.dev/github.com/zcong1993/kit) [![Go Report Card](https://goreportcard.com/badge/github.com/zcong1993/kit)](https://goreportcard.com/report/github.com/zcong1993/kit)

> go kit for microservice

## Features

- logger zap
- metrics prometheus
- tracing opentelemetry
- breaker go-zero
- shedder go-zero
- health prober
- pprof

All features can be controlled passing command line parameters.

```bash
$ go run ./_examples/gin/main.go -h

Usage:
   [flags]

Flags:
      --breaker.disable               If disable breaker
  -h, --help                          help for this command
      --inner.addr string             Inner metrics/pprof http server addr (default ":6060")
      --inner.disable                 If disable inner http server
      --inner.grace-period duration   Inner http exit grace period
      --inner.log-control             If enable logger level control, GET PUT /log/level
      --inner.with-metrics            If expose metrics router, /metrics (default true)
      --inner.with-pprof              If enable pprof routes, /debug/pprof/*
      --log.caller                    If with caller field
      --log.encoding string           Zap log encoding, console json or others you registered
      --log.level string              Log level, one of zap level (default "info")
      --log.prod                      If use zap log production preset
      --shedder.cpu-threshold int     CpuThreshold config for adaptive shedder, set 0 to disable (default 900)
```

## Examples

- [_examples/gin](./_examples/gin) simple single gin service
- [_examples/mono](./_examples/mono) multi gin services
- [_examples/grpc](./_examples/grpc) two grpc relay services and gin service as gateway

## License

MIT &copy; zcong1993
