python-reliable-processing
==========================

In this example we build a reliable event processor, filtering and normalizing events produced by the `python-gdax-connector` and pushing them in an output stream.

To achieve reliability we store in each output record the position we found the record at in the source stream. This allows us to restart from exactly the right position in cas of restart of crash, by fetching the last record from the output stream and extracting the stored position.

This strategy can be applied with any output storage, not only Styx, as long as both the processed record and position are stored atomically (either in the same row, document or record, or through a transaction). As Styx stores records atomically and durably, this gives us reliable (or exactly-once) processing.

Usage
-----

Install requirements

```bash
pip3 install -r requirements.txt
```

Ensure Styx is running and create a "matches" log

```bash
curl localhost:8000/logs -X POST -d name=matches
```

Run the processor

```bash
python3 main.py
```

From another terminal, check that the log is correctly filling with processed events

```bash
styx logs read matches --follow
```

To stress test the system, try interrupting or killing the processor and checking that it restarts from the correct position, neither dropping nor duplicating events.
