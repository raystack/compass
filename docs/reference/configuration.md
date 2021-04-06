# Configurations

This page contains reference for all the application configurations for Columbus.

## Table of Contents

* [Generic](configuration.md#-generic)
* [Search](configuration.md#-search)
* [Lineage](configuration.md#-lineage)
* [Telemetry](configuration.md#-telemetry)

## Generic

Columbus's required variables to start using it.

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

### `LOG_LEVEL`

* Example value: `error`
* Type: `optional`
* Default: `info`
* Logging level, can be one of `trace`, `debug`, `info`, `warning`, `error`, `fatal`, `panic`.

## Search

Variables for Columbus's Search or Discovery feature.

### `SEARCH_WHITELIST`

* Example value: `topic,bqtable`
* Type: `optional`
* Default `""`
* List of types that will be searchable. leave it empty if you want to run search on everything

### `SEARCH_TYPES_CACHE_DURATION`

* Example value: `1000`
* Type: `optional`
* Default `"300"`
* Duration for `Type` list cached in search service for performance purposes.
* Setting this to `0` will disable the cache which would fetch types to storage for every search request.

## Lineage

Variables for Columbus's Lineage feature.

### `LINEAGE_REFRESH_INTERVAL`

* Example value: `24h`
* Type: `optional`
* Default value: `5m`
* Refresh interval for lineage.

## Telemetry

Variables for metrics gathering.

### `STATSD_ADDRESS`

* Example value: `127.0.0.1:8125`
* Type: `optional`
* statsd client to send metrics to.

### `STATSD_PREFIX`

* Example value: `discovery`
* Type: `optional`
* Default: `columbusApi`
* Prefix for statsd metrics names.

### `STATSD_ENABLED`

* Example value: `true`
* Type: `required`
* Default: `false`
* Enable publishing application metrics to statsd.

### `NEW_RELIC_APP_NAME`

* Example value: `columbus-integration`
* Type: `optional`
* Default: `columbus`
* New Relic application name.

### `NEW_RELIC_LICENSE_KEY`

* Example value: `mf9d13c838u252252c43ji47q1u4ynzpDDDDTSPQ`
* Type: `optional`
* Default: `""`
* New Relic license key. Empty value would disable newrelic monitoring.

