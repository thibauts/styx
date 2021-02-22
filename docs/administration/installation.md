Install
-------

### Downloading a release build

Download release build

```bash
$ TODO Download release 
```

### Building from source

Building from source requires golang and git installed on the target system. 

Clone repository

```bash
$ TODO git clone public repository
```

Build
```bash
$ cd styx
$ go build -o styx-server cmd/styx-server/main.go 
```

### Running Styx

Setup data directory

```bash
$ mkdir data
```

Run styx

```bash
$ styx-server --config ./config.toml --log-level TRACE
```

### Running Styx with Docker

Build Image

```bash
$ docker build -t styx .
```

Run

```bash
$ docker run -it --rm -p 8000:8000 --name styx styx
```

Run using host data directory

```bash
$ docker run -it --rm -p 8000:8000 --mount type=bind,source="$(pwd)"/data,target=/data --name styx styx
```