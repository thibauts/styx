# Styx

Styx is a simple and high-performance event streaming platform. It aims to provide teams of all sizes with a simple to operate, disk persisted publish-subscribe system for event streams. Styx is deployed as a single binary with no dependencies, it exposes a high-performance binary protocol as well as HTTP and WebSockets APIs for both event production and consumption.

Designed around the concept of a Commit Log as popularized by projects like Apache Kafka or Apache Pulsar, Styx provides durable storage and atomicity on tail-able event logs.

Main features:

- Deployed as a **single binary** with **no dependency**
- Designed to run and provide durability as a **single-node** system
- Out-of-the-box **Prometheus** and **Statsd** monitoring
- **Full-featured CLI** for event log management, **backup/restore** and flexible **bulk load/export**
- Event records are **immutable**, **atomic**, **durable**, and fsynced to permanent storage before being acked
- Configurable **retention policies**
- **Small footprint**. Very low memory and CPU consumption
- Native **REST API** for event log **management**
- Native **REST API** for event **production and consumption**, including batched with long-polling support
- Native **WebSockets API** both for event production and consumption
- **Millions of events/second**, **GB/s throughput**, low latency
- Scales to **thousands of producers, consumers, and event logs**
- **Decoupled storage engine** that can be used standalone

Styx is designed to be a no-brainer for teams that don't want to invest resources in complex (and sometimes fragile) clustered solutions, while providing data safety, plenty of room to scale and instant integration in any language.

## How does it work ?

Styx is a Publish-Subscribe system. `Producers` publish events to event `logs`, and `consumers` subscribe to these logs for further processing.

```
                                           +--------------+
                                           |              |
                                     +---->+   Consumer   |
                                     |     |              |
                                     |     +--------------+
+------------+     +------------+    |     +--------------+
|            |     |            +----+     |              |
|  Producer  +---->+  Styx log  +--------->+   Consumer   |
|            |     |            +----+     |              |
+------------+     +------------+    |     +--------------+
                                     |     +--------------+
                                     |     |              |
                                     +---->+   Consumer   |
                                           |              |
                                           +--------------+
```


Usage of the word `log` in Styx should not be confused with textual application logs, and designate an ordered sequence of event `records` that are persisted to disk.

```
+------------+------------+------------+-----------+------------+
|  record 0  |  record 1  |  record 2  |  .......  |  record N  |
+------------+------------+------------+-----------+------------+
```

Event `logs` are append-only and immutable. As they are persisted to disk, event `logs` can be consumed in real-time by susbcribers, and can also be replayed from the beggining or from any arbitrary `position`. This unlocks keeping a complete history of published events, regenerating downstream data when updating your app, or replaying complete streams through new processing logic as unexpected needs surface.

## Using Styx

Styx usage can differ from traditional Message Queues and Streaming Platforms. In particular we encourage usage patterns that remove the need for consumer acks while providing stronger guarantees. You should consider reading the [Reliable Processing Howto](./docs/howto/reliable-processing.md) as an introduction.

Reliable event production from external sources to Styx logs can also be achieved with data sources exhibiting particular properties. We plan on developing this subject and other usage patterns in Howto's and tutorial as soon as possible.

## Getting Started

### Using docker

You can launch a Styx container to try it out with

```bash
$ docker run --name styx -d -p 7123:7123 dataptive/styx
```

Styx will now be reachable at http://localhost:7123/.

The Styx CLI is available from the docker container

```bash
$ docker exec -it styx styx
Usage: styx COMMAND

A command line interface (CLI) for the Styx API.

Commands:
        logs  Manage logs

Global Options:
        -f, --format string     Output format [text|json] (default "text")
        -H, --host string       Server to connect to (default "http://localhost:7123")
        -h, --help              Display help
```

### Installing or building from source

To build or install Styx from source, you will need to have a working Go environment.

```bash
$ go get github.com/dataptive/styx/cmd/styx-server
$ go get github.com/dataptive/styx/cmd/styx
```

Create a directory to host event log data, and update your `config.toml` to make `data_directory` point to your newly created directory (a sample config file is provided at the root of this repository).

Assuming you're hosting both `data_directory` and `config.toml` in your current working directory, run

```bash
$ mkdir $PWD/data
```

Then run the Styx server with this command

```bash
$ styx-server --config=$PWD/config.toml
```

The Styx CLI should be available as the `styx` command

```bash
$ styx
Usage: styx COMMAND

A command line interface (CLI) for the Styx API.

Commands:
        logs  Manage logs

Global Options:
        -f, --format string     Output format [text|json] (default "text")
        -H, --host string       Server to connect to (default "http://localhost:7123")
        -h, --help              Display help
```

To build the binaries without installing them, run

```bash
$ git clone https://github.com/dataptive/styx.git
$ cd styx
$ go build -o styx-server cmd/styx-server/main.go
$ go build -o styx cmd/styx/main.go
```

The `styx-server` and `styx` binaries should be available in your working directory.

### Producing and consuming records

While the server is running, create a `test` event log and tail its contents using the CLI

```bash
$ styx logs create test
$ styx logs read test --follow
```

The CLI will hang, waiting for incoming events (you can quit with Ctrl+D)

In another terminal, try pushing events in the log with curl

```bash
$ curl localhost:7123/logs/test/records -X POST -d 'Hello, world !' 
```

All CLI commands support the `--help` flag, play with them to get a quick tour of Styx's features. For example

```bash
$ styx logs create --help
Usage: styx logs create NAME [OPTIONS]

Create a new log

Options:
        --max-record-size bytes         Maximum record size
        --index-after-size bytes        Write a segment index entry after every size
        --segment-max-count records     Create a new segment when current segment exceeds this number of records
        --segment-max-size bytes        Create a new segment when current segment exceeds this size
        --segment-max-age seconds       Create a new segment when current segment exceeds this age
        --log-max-count records         Expire oldest segment when log exceeds this number of records
        --log-max-size bytes            Expire oldest segment when log exceeds this size
        --log-max-age seconds           Expire oldest segment when log exceeds this age

Global Options:
        -f, --format string             Output format [text|json] (default "text")
        -H, --host string               Server to connect to (default "http://localhost:7123")
        -h, --help                      Display help
```

## Documentation

The documentation is a work in progress, please open an issue if you find something unclear or missing.

* [Getting started](./docs/administration)
	* [Installing](./docs/administration/installation.md)
	* [Configuration](./docs/administration/configuration.md)
	* [Monitoring with Prometheus and Statsd](./docs/administration/monitoring.md)
	* [CLI reference](./docs/administration/CLI.md)
* [API reference](./docs/api)
	* [Managing event logs](./docs/api/manage.md)
	* [Producing events through the REST API](./docs/api/write_HTTP.md)
	* [Consuming events through the REST API](./docs/api/read_HTTP.md)
	* [Producing events with WebSockets](./docs/api/write_WS.md)
	* [Consuming events with WebSockets](./docs/api/read_WS.md)
	* [Binary protocol specs](./docs/api/styx_protocol.md)
		* [Opening a connection for producing events](./docs/api/write_styx.md)
		* [Opening a connection for consuming events](./docs/api/read_styx.md)
	* [API Media Types](./docs/api/media_types.md)
	* [API errors](./docs/api/errors.md)
* [Howto](./docs/howto)
	* [Consuming events with HTTP long-polling](./docs/howto/longpolling.md)
* [Code examples](./docs/examples)
	* [Using the Golang client Producer](./docs/examples/golang-fast-producer)
	* [Using WebSockets in python to consume a real-time event feed](./docs/examples/python-gdax-connector)
	* [Using python+requests to produce concurrently](./docs/examples/python-multi-producers)
	* [Building a reliable processing pipeline using python and WebSockets](./docs/examples/python-reliable-processing)
	* [Building a HTTP->Styx event gateway in NodeJS](./docs/examples/nodejs-event-gateway)
	* [Using NodeJS to build a simple event processing pipeline](./docs/examples/nodejs-simple-processing)

## Design

TODO !

## License

Styx is published under an Apache 2.0+BSL [LICENSE](LICENSE). The BSL part is meant to prevent cloud providers from selling Styx as a service at zero cost. Apart from that, you should be safe to use it as you see fit, for example as a service inside your company, or even as a value-added service in a product you are distributing or selling.

## Contributing

You're more than welcome to open issues should you encounter any bug or instability in the code. Feature suggestions are welcome. We may restrict PRs until we have setup a CLA.

## Community

If you wish to join our Slack, request an invite at hello@dataptive.io

