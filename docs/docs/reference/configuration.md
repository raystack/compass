# Configurations

This page contains reference for all the application configurations for Compass.

## Table of Contents

* [Generic](configuration.md#-generic)
* [Telemetry](configuration.md#-telemetry)

## Generic

Compass's required variables to start using it.
### `LOG_LEVEL`

* Example value: `error`
* Type: `optional`
* Default: `info`
* Logging level, can be one of `trace`, `debug`, `info`, `warning`, `error`, `fatal`, `panic`.
### `SERVER_HOST`

* Example value: `localhost`
* Type: `required`
* Network interface to bind to.

### `SERVER_PORT`

* Example value: `8080`
* Type: `required`
* Port to listen on.

### `ELASTICSEARCH_BROKERS`

* Example value: `http://localhost:9200,http://localhost:9300`
* Type: `required`
* Comma separated list of elasticsearch nodes.
### `DB_HOST`
* Example value: `localhost`
* Type: `required`
* PostgreSQL DB hostname to connect.
### `DB_PORT`
* Example value: `5432`
* Type: `required`
* PostgreSQL DB port to connect.
### `DB_NAME`
* Example value: `compass`
* Type: `required`
* PostgreSQL DB name to connect.
### `DB_USER`
* Example value: `postgres`
* Type: `required`
* PostgreSQL DB user to connect.
### `DB_PASSWORD`
* Example value: `~`
* Type: `required`
* PostgreSQL DB user's password to connect.
### `DB_SSL_MODE`
* Example value: `disable`
* Type: `optional`
* PostgreSQL DB SSL mode to connect.
### `IDENTITY_UUID_HEADER`
* Example value: `Compass-User-UUID`
* Type: `required`
* Header key to accept Compass User UUID. See the API reference for more information.
### `IDENTITY_EMAIL_HEADER`
* Example value: `Compass-User-Email`
* Type: `optional`
* Header key to accept Compass User Email. See the API reference for more information.
### `IDENTITY_PROVIDER_DEFAULT_NAME`
* Example value: `shield`
* Type: `optional`
* Default value of user provider. See the API reference for more information.

## Telemetry

Variables for metrics gathering. Compass uses OpenTelemetry for traces and metrics.

### `TELEMETRY_SERVICENAME`

* Example value: `compass`
* Type: `optional`
* Default: `compass`
* Service name reported to the OpenTelemetry collector.

### `TELEMETRY_OPENTELEMETRY_ENABLED`

* Example value: `true`
* Type: `optional`
* Default: `false`
* Enable OpenTelemetry traces and metrics.

### `TELEMETRY_OPENTELEMETRY_COLLECTORADDR`

* Example value: `localhost:4317`
* Type: `required` (when enabled)
* gRPC address of the OpenTelemetry collector.

### `TELEMETRY_OPENTELEMETRY_TRACESAMPLEPROBABILITY`

* Example value: `0.1`
* Type: `optional`
* Default: `1.0`
* Trace sampling probability (0.0 to 1.0).

