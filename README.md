# cron-me

Name subject to change

## Prometheus
```
PUSHGATEWAY=http://pushgateway.example/jobs/cron/
@hourly cron-me -- /path/to/script

TEXTFILE=$HOME/prom/output.prom
@daily cron-me -- /path/to/script
```

## Growl
```
GROWL=127.0.0.1
@hourly cron-me -- /path/to/script
```

## Sentry
```
SENTRY_DSN=http://sentry.example/
@hourly cron-me -- /path/to/script
```
