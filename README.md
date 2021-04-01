# Styx

Styx is a simple and high-performance event streaming broker. It aims to provide teams of all sizes with a simple to operate, disk-persisted publish-subscribe system for event streams. Styx is deployed as a single binary with no dependencies, it exposes a high-performance binary protocol as well as HTTP and WebSockets APIs for both event production and consumption.

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
$ docker pull dataptive/styx
$ docker run --name styx -p 7123:7123 dataptive/styx
2021-03-31T08:06:06.987817093Z INFO server: starting Styx server version 0.1.4
2021-03-31T08:06:06.987971158Z INFO logman: starting log manager (data_directory=./data)
2021-03-31T08:06:06.988911129Z INFO server: listening for client connections on 0.0.0.0:7123
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

### Precompiled binaries

Precompiled binaries and packages are available from the [Releases](https://github.com/dataptive/styx/releases) section for various operating systems and architectures.

Installing on Debian-based systems

```bash
wget https://github.com/dataptive/styx/releases/download/v0.1.4/styx-0.1.4-amd64.deb
dpkg -i styx-0.1.4-amd64.deb
```

Installing on Redhat-based systems

```bash
wget https://github.com/dataptive/styx/releases/download/v0.1.4/styx-0.1.4-amd64.rpm
rpm -i styx-0.1.4-amd64.rpm
```

### Building from source

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
$ styx logs consume test --follow
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

### Benchmarking

Use the `benchmark` command from the CLI to run a benchmark on your own server.

The command will launch a batch of produce and consume tasks with various payload sizes and produce a report. You should have about 10GB of free storage for everything to run smoothly. Used disk space will automatically be freed when the benchmark ends.

Although you'll get better numbers when running outside of docker, the following commands will give you a quick estimate of what you can expect from Styx in your own context.

```bash
$ docker pull dataptive/styx
$ docker run --name styx -d -p 7123:7123 dataptive/styx
$ docker exec styx styx benchmark
```

In case you've built Styx from source, simply run

```bash
$ styx benchmark
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
	* [Producing events through the REST API](./docs/api/produce_HTTP.md)
	* [Consuming events through the REST API](./docs/api/consume_HTTP.md)
	* [Producing events with WebSockets](./docs/api/produce_websocket.md)
	* [Consuming events with WebSockets](./docs/api/consume_websocket.md)
	* [Binary protocol specs](./docs/api/styx_protocol.md)
		* [Opening a connection for producing events](./docs/api/produce_styx.md)
		* [Opening a connection for consuming events](./docs/api/consume_styx.md)
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

## Contributing

You're more than welcome to open issues should you encounter any bug or instability in the code. Feature suggestions are welcome. We may restrict PRs until we have setup a CLA.

## Community

You're welcome to join our [Slack](https://join.slack.com/t/dataptive/shared_invite/zt-mzl99jf9-zaFiIVYbZRN6m2nqkcgZmg) to discuss the project or ask for help !

