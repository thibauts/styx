Styx protocol
-------------

## Motivation

Styx provides HTTP endpoints to read and write records but it also provide a binary stream protocol for situations where high speed and low latencies are a concern.

This protocol is based on TCP and is message oriented.
The protocol has two parts, an handshake using HTTP and a data stream transfer.
We use the HTTP upgrade mechanism for handshake as described in [rfc 2616](https://tools.ietf.org/html/rfc2616#section-14.42).

This mecanism has many advantages.
It allows us to rely on HTTP for TLS, authentication etc.
If an error occurs for some reason, we can return JSON HTTP errors before starting the stream.
Handshake complexity is offloaded to HTTP, making the the protocol simpler for client implementers.

## Description

Styx protocol has two parts: an handshake and a data transfer.

The handshake is made using an HTTP call to the server.

The raw handshake on client side:

```http
POST / HTTP/1.1
Host: server.example.com
Upgrade: styx/0
Connection: Upgrade
X-Styx-Timeout: 50
```

Note we use the method GET to signal to the server the client ask for record messages and POST when the client will provide record messages. 

Also note that additional query parameters can influence the nature of the data transfer part. See Styx records endpoints for more about it.

The raw handshake from server side:
```http
HTTP/1.1 101 Switching Protocols
Upgrade: styx/0
Connection: Upgrade
X-Styx-Timeout: 40
```

The `X-Styx-Timeout` header value in seconds is the maximum amount of time the peer will keep the connection opened whitout receiving messages.
Each peer should periodically send heartbeat messages if no others messages are sent.
The value of this period must be significantly lower than `X-Styx-Timeout` to keep the TCP connection alive.


When both peers have received their handshake and if it was successful, the data transfer on the TCP connection can start using messages.


## Messages

Messages are used during the data transfert part of Styx protocol.

All integer values of the protocol are big-endian ordered.

| Type      | Code (int16) |
| ----------| -------------|
| Record    | 1            |
| Ack       | 2            | 
| Heartbeat | 3            |
| Error     | 4            |

### Record message

Record messages are used for transmitting records between two peers.

```
  +----------------+--------------------------------+--------------------------------+
  |  type (int16)  |          size (int32)          |       record (size bytes)      |
  +----------------+--------------------------------+--------------------------------+
```

### Ack message

Ack messages are sent back from the server to the client to confirm log records were syncronized on disk.
Since Styx protocol is message stream oriented and not request response oriented, a ack message is not expected for every record message. 

```
  +----------------+--------------------------------+--------------------------------+
  |  type (int16)  |        position (int64)        |           count (int64)        |
  +----------------+--------------------------------+--------------------------------+
```

`count` keeps track of the number of records sent and successfully synchronized on disk since the beginning of the stream.  
`position` contains the highest syncronized position in the log.


### Heartbeat message

Heartbeat messages are sent from both peers to signal they are still alive. 

```
  +----------------+
  |  type (int16)  |
  +----------------+
```

### Error message

Error messages are used to signal an unrecoverable error happened and thus the stream must come to an end.

```
  +----------------+----------------+
  |  type (int16)  |  code (int16)  |
  +----------------+----------------+
```

`code` contains an error code adding precision about what happened. The value for an unknwon error is `0`.
