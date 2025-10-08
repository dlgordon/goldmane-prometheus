This is a Go application that will run as a small daemon in Kubernetes, which should connect to the Calico Goldmane API (https://docs.tigera.io/calico/latest/observability/flow-logs-api) and present a /metrics endpoint that will export Prometheus metrics that can be used to track success and failed flow metrics.

The /metrics endpoint should expose 2 metrics, which are of type counter

calico_flow_allow
calico_flow_deny

For each metric, the following dimensions should be tracked.

- Reporter (Src vs dst)
- Protocol
- Src namespace
- Src pod
- Src port
- Dst namespace
- Dst object
- Dst port

The Goldmane API is a gRPC based API - the proto file is available at https://github.com/projectcalico/calico/blob/master/goldmane/proto/api.proto

Bootstrap a basic Go application that can sit in a loop, process the data and expose a /metrics endpoint. The processing should happen on a a configurable period, by default 15 seconds. You should write idiotmatic Go code, drawing on examples of similar Prometheus generators.