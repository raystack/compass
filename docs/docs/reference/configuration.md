# Configurations

This page contains reference for all the application configurations for Compass.

## Table of Contents

* [Generic](configuration.md#generic)
* [Database](configuration.md#database)
* [Service](configuration.md#service)
* [Telemetry](configuration.md#telemetry)

## Generic

### `LOG_LEVEL`

* Example value: `error`
* Type: `optional`
* Default: `info`
* Logging level, can be one of `trace`, `debug`, `info`, `warning`, `error`, `fatal`, `panic`.

## Database

Compass uses PostgreSQL (with pgvector and pg_trgm extensions) as its sole storage.

### `DB_HOST`
* Example value: `localhost`
* Type: `required`
* PostgreSQL DB hostname to connect.
### `DB_PORT`
* Example value: `5432`
* Type: `required`
* Default: `5432`
* PostgreSQL DB port to connect.
### `DB_NAME`
* Example value: `compass`
* Type: `required`
* Default: `postgres`
* PostgreSQL DB name to connect.
### `DB_USER`
* Example value: `compass`
* Type: `required`
* Default: `root`
* PostgreSQL DB user to connect.
### `DB_PASSWORD`
* Example value: `~`
* Type: `required`
* PostgreSQL DB user's password to connect.
### `DB_SSLMODE`
* Example value: `disable`
* Type: `optional`
* Default: `disable`
* PostgreSQL DB SSL mode to connect.

## Service

### `SERVICE_HOST`

* Example value: `localhost`
* Type: `optional`
* Default: `0.0.0.0`
* Network interface to bind to.

### `SERVICE_PORT`

* Example value: `8080`
* Type: `optional`
* Default: `8080`
* Port to listen on.

### `SERVICE_BASEURL`

* Example value: `localhost:8080`
* Type: `optional`
* Default: `localhost:8080`
* Base URL for the server.

### `SERVICE_IDENTITY_HEADERKEY_UUID`
* Example value: `Compass-User-UUID`
* Type: `optional`
* Default: `Compass-User-UUID`
* Header key to accept Compass User UUID.

### `SERVICE_IDENTITY_HEADERKEY_EMAIL`
* Example value: `Compass-User-Email`
* Type: `optional`
* Default: `Compass-User-Email`
* Header key to accept Compass User Email.

### `SERVICE_IDENTITY_PROVIDER_DEFAULT_NAME`
* Example value: `shield`
* Type: `optional`
* Default value of user provider.

### `SERVICE_IDENTITY_NAMESPACE_CLAIM_KEY`
* Example value: `namespace_id`
* Type: `optional`
* Default: `namespace_id`
* JWT claim key used to extract the namespace ID.

### `SERVICE_CORS_ALLOWED_ORIGINS`
* Example value: `*`
* Type: `optional`
* Default: `*`
* Comma-separated list of allowed CORS origins.

### `SERVICE_MAX_RECV_MSG_SIZE`
* Example value: `33554432`
* Type: `optional`
* Default: `33554432` (32MB)
* Maximum receive message size in bytes.

### `SERVICE_MAX_SEND_MSG_SIZE`
* Example value: `33554432`
* Type: `optional`
* Default: `33554432` (32MB)
* Maximum send message size in bytes.

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
