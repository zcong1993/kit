<a name="unreleased"></a>
## [Unreleased]


<a name="v0.2.1"></a>
## [v0.2.1] - 2021-06-30
### Docs
- tweak badges
- add ci badges
- add badges
- **changelog:** add changelog generator


<a name="v0.2.0"></a>
## [v0.2.0] - 2021-06-30
### Chore
- **breaker:** add comments
- **extapp:** add comments
- **extflag:** add comments
- **extgrpcc:** add comments
- **ginhelper:** add comments
- **lint:** tweak golang ci lint config
- **log:** add comments
- **metrics:** add comments
- **prober:** remove nop prober
- **runutil:** add comments
- **server:** add comments
- **shedder:** add comments
- **tracing:** add comments
- **zero:** add comments

### Ci
- use correct lint version


<a name="v0.1.0"></a>
## v0.1.0 - 2021-06-30
### Chore
- add comment
- fix deprecated functions
- bump deps
- split _example from x package
- remove opentracing
- bump grpc
- bump deps
- refactor shedder breaker register and tweak examples folder
- tweak logger
- tweak name
- add otel mono example
- add otel grpc example
- tweak
- update default server
- add on shutdown hook
- never use prometheus.DefaultRegisterer
- tweak flag help text
- **breaker:** add logger
- **breaker:** clean code
- **example:** add comments
- **examples:** add mono services example
- **examples:** add grpc server
- **examples:** grpc
- **examples:** tweak
- **examples:** update
- **examples:** update example
- **extapp:** add fatal helpers
- **extapp:** tweak
- **exthttp:** move http server to server/exthttp
- **lint:** tweak ci lint
- **log:** tweak log flag register
- **logger:** remove ua
- **metrics:** reg.MustRegister once
- **mux:** mux expose HandleFunc and Handle
- **tracing:** set force tracing  baggage item only when ForceTracingBaggageKey header is set
- **tracing:** rename MustSetupTracer -> MustInitTracer
- **tracing:** update gin example
- **zero:** disable go-zero log

### Ci
- bump deps
- tweak build
- fix actions
- tweak build config
- **actions:** cache go mod
- **build:** fix build action
- **build:** fix build action
- **lint:** add build test
- **lint:** add golang ci lint action

### Docs
- update docs

### Feat
- add NewFromFlags logger constructor
- allow disable breaker
- add Run for ext server
- add otel tracing
- add log debug helper
- update ginhelper error
- add logger factory
- add gin error handler wrapper
- auto response validate error
- **breaker:** add breaker
- **breaker:** add grpc server breaker
- **extapp:** add inner http server helper
- **extapp:** add extapp default http server
- **extapp:** return inner http server
- **extflag:** add path of content flag helper
- **log:** add inner log control route
- **logger:** add logger
- **metrics:** add grpc metrics
- **metrics:** add http metrics instrumentation middleware
- **metrics:** add grpc client metrics
- **prober:** add prober
- **runutil:** recover
- **runutil:** run utils
- **shedder:** add shedder
- **shedder:** add grpc server shedder middleware
- **tracing:** add grpc client tracing
- **tracing:** add grpc server tracing
- **tracing:** add tracing
- **tracing:** add HTTPAutoTripperware

### Fix
- import
- **tracing:** fix force tracing header key not works
- **tracing:** fix response trace-id

### Refactor
- **breaker:** move package
- **log:** use cobra


[Unreleased]: https://github.com/zcong1993/kit/compare/v0.2.1...HEAD
[v0.2.1]: https://github.com/zcong1993/kit/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/zcong1993/kit/compare/v0.1.0...v0.2.0
