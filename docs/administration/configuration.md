Configuration
-------------

At startup Styx looks for a `config.toml` in the working directory.  
Config path can be overridden using Styx server `--config` option.

```bash
$ styx-server --config /etc/styx/config.toml
```

### Global settings

| Setting                        | Description                                                                                       |
|--------------------------------|---------------------------------------------------------------------------------------------------|
| `pid_file`                     | Path for Styx pid file.                                                                           |
| `bind_address`                 | Address Styx will bind to.                                                                        |
| `shutdown_timeout`             | Number of seconds before triggering a hard shutdown when Styx receives a SIGINT or SIGTERM signal.|
| `cors_allowed_origins`         | An array of allowed origins. `"*"` Allow all origins.                                             |
| `http_read_buffer_size`        | Size of Styx internal read buffers over HTTP.                                                     |
| `http_write_buffer_size`       | Size of Styx internal write buffers over HTTP.                                                    |
| `tcp_read_buffer_size`         | Size of Styx internal read buffers over TCP.                                                      |
| `tcp_write_buffer_size`        | Size of Styx internal write buffers over TCP.                                                     |
| `tcp_timeout`                  | Number of seconds before shutting down a Styx Protocol connection when idle.                      |

### Log manager settings

**[log_manager]**

| Setting             | Description                         |
|---------------------|-------------------------------------|
| `data_directory`    | Path for Styx logs storage.         |
| `write_buffer_size` | Size of internal log writer buffer. |

### Metrics

**[metrics.statsd]**

| Setting   | Description                 |
|-----------|-----------------------------|
| `address` | Address of statsd server.   |
| `prefix`  | Statsd log metrics prefix.  |
