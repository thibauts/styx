Monitor
-------

### Prometheus

By default Styx provides an endpoint with Prometheus metrics.

```bash
$ curl --XGET 'http://localhost:8000/metrics'
```

The response provide various information about Styx process, such as number of goroutines memory etc.

Logs informations are also provided.

```
# HELP log_file_size Current log file size
# TYPE log_file_size gauge
log_file_size{log="myLog"} 487
# HELP log_record_count Current record count
# TYPE log_record_count gauge
log_record_count{log="myLog"} 60
```

### Statsd

Log Metrics can also be reported to a Statsd server when enabled in the Styx [config](./configuration.md).

Log names are added to the metric path as follow.

```
log.myLog.file.size487|g
log.myLog.record.count60|g
```
