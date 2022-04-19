# Telemetry

Compass collects basic HTTP metrics (response time, duration, etc) and sends it to [statsd](https://github.com/statsd/statsd) and [New Relic](https://newrelic.com/) when enabled.

## Statsd
By default statsd is not enabled. To enable statsd, we just need to set these configurations below

```
STATSD_ENABLED=true
STATSD_ADDRESS=127.0.0.1:8125
STATSD_PREFIX=compass
```


## New Relic
Similar with statsd, New Relic is not enabled by default. To enable New Relic, you can set these configurations

```
NEW_RELIC_LICENSE_KEY=mf9d13c838u252252c43ji47q1u4ynzpDDDDTSPQ
NEW_RELIC_APP_NAME=compass
```

Empty `NEW_RELIC_LICENSE_KEY` will disable New Relic integration.
