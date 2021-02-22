nodejs-event-gateway
====================

In this example, we build a simple tracking event gateway with raw NodeJS and Styx.

Usage
-----

Ensure Styx is running and create a "events" log using the Styx CLI

```bash
styx logs create events
```

Run the gateway

```bash
node index.js
```

From another terminal, follow the log with the Styx CLI

```bash
styx logs read events --follow
```

From yet another terminal, send requests to the gateway on port 9000, varying query params

```bash
curl "localhost:9000/?type=view&payload=hello"
curl "localhost:9000/?type=view&payload=styx"
```
