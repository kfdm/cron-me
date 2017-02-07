# cron-me

Name subject to change

## Prometheus
```
PUSHGATEWAY=http://pushgateway.example/jobs/cron/
@hourly cron-me -- /path/to/script

TEXTFILE=$HOME/prom/output.prom
@daily cron-me -- /path/to/script
```

```
# HELP cron_last_run_unixtimestamp Last run time as a unix timestamp
cron_last_run_unixtimestamp 1234567
# HELP cron_last_run_duration Last run time in seconds
cron_last_run_duration_seconds 123
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
