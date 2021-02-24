## Reliable processing

As opposed to traditional Message Queueing or Streaming Platforms like RabbitMQ, Apache Kafka or Apache Pulsar, Styx does not provide a message acknowledgment mecanism. Message ACKing can still be implemented by creating consumer `position` logs with limited retention, but we encourage patterns that in our opinion are both simpler and safer to use.

Relying on the atomicity guarantees provided by Styx (an event `record` is either completely persisted or is not persisted at all), the recommended pattern to keep track of consumer `position` is to store it atomically with consumer output data in the `target` log.

A consumer implementing a counter would keep track of its state like this:

Input event #123, read from the `source` log.

```json
{
	"type": "increment",
	"value": 1
}
```

Output state written by the consumer to the `target` log.

```json
{
	"total": 19872,
	"position": 123
}
```

In case of failure of either the consumer or Styx, the consumer can easily find its restart position by requesting the last event record from the `target` log, extracting the `position` field and retarting consumption from `position+1`. This avoids the *at-least-once* and *at-most-once* issues in processing setups. See [python-reliable-processing](../examples/python-reliable-processing) for an example Python script implementing this pattern.

This strategy can be implemented with any output Store or Database, as long as the consumer `position` can be stored atomically with the output data, either in the same row, document, record, or in a transaction.
