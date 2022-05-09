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
### DB_HOST
* Example value: `localhost`
* Type: `required`
* PostgreSQL DB hostname to connect.
### DB_PORT
* Example value: `5432`
* Type: `required`
* PostgreSQL DB port to connect.
### DB_NAME
* Example value: `compass`
* Type: `required`
* PostgreSQL DB name to connect.
### DB_USER
* Example value: `postgres`
* Type: `required`
* PostgreSQL DB user to connect.
### DB_PASSWORD
* Example value: `~`
* Type: `required`
* PostgreSQL DB user's password to connect.
### DB_SSL_MODE
* Example value: `disable`
* Type: `optional`
* PostgreSQL DB SSL mode to connect.
### IDENTITY_UUID_HEADER
* Example value: `Compass-User-UUID`
* Type: `required`
* Header key to accept Compass User UUID. See [User](../concepts/user.md) for more information about the usage.
### IDENTITY_EMAIL_HEADER
* Example value: `Compass-User-Email`
* Type: `optional`
* Header key to accept Compass User Email. See [User](../concepts/user.md) for more information about the usage.
### IDENTITY_PROVIDER_DEFAULT_NAME
* Example value: `shield`
* Type: `optional`
* Default value of user provider. See [User](../concepts/user.md) for more information about the usage.

## Telemetry

Variables for metrics gathering.

### `STATSD_ADDRESS`

* Example value: `127.0.0.1:8125`
* Type: `optional`
* statsd client to send metrics to.

### `STATSD_PREFIX`

* Example value: `discovery`
* Type: `optional`
* Default: `compassApi`
* Prefix for statsd metrics names.

### `STATSD_ENABLED`

* Example value: `true`
* Type: `required`
* Default: `false`
* Enable publishing application metrics to statsd.

### `NEW_RELIC_APP_NAME`

* Example value: `compass-integration`
* Type: `optional`
* Default: `compass`
* New Relic application name.

### `NEW_RELIC_LICENSE_KEY`

* Example value: `mf9d13c838u252252c43ji47q1u4ynzpDDDDTSPQ`
* Type: `optional`
* Default: `""`
* New Relic license key. Empty value would disable newrelic monitoring.

