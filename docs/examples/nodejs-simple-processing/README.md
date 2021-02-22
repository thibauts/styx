nodejs-simple-processing
========================

This examples relies on the data produced through the `nodejs-event-gateway` example. It consumes the "events" log, computes aggregates and pushes them to the "stats" log in real time.

Usage
-----

Install dependencies

```bash
npm install
```

Run the processor

```bash
node index.js
```

From another terminal, with `nodejs-event-gateway` running, simulate events on the gateway

```bash
curl "localhost:9000/?type=view"
curl "localhost:9000/?type=view"
curl "localhost:9000/?type=click"
curl "localhost:9000/?type=sale"
```

From yet another terminal, follow the contents of the "stats" log with Styx CLI

```bash
styx logs read stats --follow
```

You should see stat counters update as you send requests.

```json
{"view":1}
{"view":2}
{"view":2,"click":1}
{"view":2,"click":1,"sale":1}
```

The latest values can read by fetching the last record from the stats log

```bash
curl "localhost:8000/logs/stats/records?whence=end&position=-1" -s | jq .
``