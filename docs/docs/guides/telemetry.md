# Telemetry

Compass uses [OpenTelemetry](https://opentelemetry.io/) to collect traces and metrics and export them via OTLP gRPC to a collector.

## OpenTelemetry

By default OpenTelemetry is not enabled. To enable it, set these configurations:

```
TELEMETRY_OPENTELEMETRY_ENABLED=true
TELEMETRY_OPENTELEMETRY_COLLECTORADDR=localhost:4317
TELEMETRY_SERVICENAME=compass
```

You can optionally configure trace sampling probability:

```
TELEMETRY_OPENTELEMETRY_TRACESAMPLEPROBABILITY=0.1
```
