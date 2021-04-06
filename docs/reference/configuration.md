# Configurations

This page contains reference for all the application configurations for Columbus.

## Table of Contents
- [Generic](#-generic)
- [Search](#-search)
- [Lineage](#-lineage)
- [Telemetry](#-telemetry)

## <a name="Generic" /> Generic
Columbus's required variables to start using it.

#### <a name="SERVER_HOST" /> `SERVER_HOST`
* Example value: `localhost`
* Type: `required`

* Network interface to bind to.

#### <a name="SERVER_PORT" /> `SERVER_PORT`
* Example value: `8080`
* Type: `required`

* Port to listen on.

#### <a name="ELASTICSEARCH_BROKERS" /> `ELASTICSEARCH_BROKERS`
* Example value: `http://localhost:9200,http://localhost:9300`
* Type: `required`

* Comma separated list of elasticsearch nodes.

#### <a name="LOG_LEVEL" /> `LOG_LEVEL`
* Example value: `error`
* Type: `optional`
* Default: `info`

* Logging level, can be one of `trace`, `debug`, `info`, `warning`, `error`, `fatal`, `panic`.

## <a name="Search" /> Search
Variables for Columbus's Search or Discovery feature.

#### <a name="SEARCH_WHITELIST" /> `SEARCH_WHITELIST`
* Example value: `topic,bqtable`
* Type: `optional`
* Default `""`

* List of types that will be searchable. leave it empty if you want to run search on everything

#### <a name="SEARCH_TYPES_CACHE_DURATION" /> `SEARCH_TYPES_CACHE_DURATION`
* Example value: `1000`
* Type: `optional`
* Default `"300"`

* Duration for `Type` list cached in search service for performance purposes.
* Setting this to `0` will disable the cache which would fetch types to storage for every search request.

## <a name="Lineage" /> Lineage
Variables for Columbus's Lineage feature.

#### <a name="LINEAGE_REFRESH_INTERVAL" /> `LINEAGE_REFRESH_INTERVAL`
* Example value: `24h`
* Type: `optional`
* Default value: `5m`

* Refresh interval for lineage.

## <a name="Telemetry" /> Telemetry
Variables for metrics gathering.

#### <a name="STATSD_ADDRESS" /> `STATSD_ADDRESS`
* Example value: `127.0.0.1:8125`
* Type: `optional`

* statsd client to send metrics to.

#### <a name="STATSD_PREFIX" /> `STATSD_PREFIX`
* Example value: `discovery`
* Type: `optional`
* Default: `columbusApi`

* Prefix for statsd metrics names.

#### <a name="STATSD_ENABLED" /> `STATSD_ENABLED`
* Example value: `true`
* Type: `required`
* Default: `false`

* Enable publishing application metrics to statsd.

#### <a name="NEW_RELIC_APP_NAME" /> `NEW_RELIC_APP_NAME`
* Example value: `columbus-integration`
* Type: `optional`
* Default: `columbus`

* New Relic application name.

#### <a name="NEW_RELIC_LICENSE_KEY" /> `NEW_RELIC_LICENSE_KEY`
* Example value: `mf9d13c838u252252c43ji47q1u4ynzpDDDDTSPQ`
* Type: `optional`
* Default: `""`

* New Relic license key. Empty value would disable newrelic monitoring.
