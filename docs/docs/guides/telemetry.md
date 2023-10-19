# Telemetry

Compass collects basic HTTP metrics (response time, duration, etc) and sends it to [New Relic](https://newrelic.com/) when enabled.


## New Relic
New Relic is not enabled by default. To enable New Relic, you can set these configurations

```
NEW_RELIC_LICENSE_KEY=mf9d13c838u252252c43ji47q1u4ynzpDDDDTSPQ
NEW_RELIC_APP_NAME=compass
```

Empty `NEW_RELIC_LICENSE_KEY` will disable New Relic integration.
