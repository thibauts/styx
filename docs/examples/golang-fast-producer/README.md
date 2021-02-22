golang-fast-producer
====================

In this example, we'll push events to Styx at a very high rate using the Styx protocol in Golang.

Usage
-----

Install dependencies

```bash
go get gitlab.com/dataptive/styx
```

Ensure Styx is running and create a "fast" log

```bash
curl localhost:8000/logs -X POST -d name=fast
```

Setup a watch on the log to monitor its growth

```bash
watch curl localhost:8000/logs/fast -s
```

In another terminal, run the producer

```bash
go run main.go
```

When you're done, it may be a good idea to free disk space by deleting the log

```bash
curl localhost:8000/logs/fast -X DELETE
```
