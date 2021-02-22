Command Line Interface
----------------------

Styx command line interface allows you to interact with Styx logs.  
To display this usage, either run `styx logs` or `styx logs --help`.


```bash
$ styx logs -h
Usage: styx logs COMMAND

Manage logs

Commands:
        list                    List available logs
        create                  Create a new log
        get                     Show log details
        delete                  Delete a log
        backup                  Backup a log
        restore                 Restore a log
        write                   Write records to a log
        read                    Read records from a log

Global Options:
        -f, --format string     Output format [text|json] (default "text")
        -H, --host string       Server to connect to (default "http://localhost:8000")
        -h, --help              Display help
```

## List logs

### Usage

```bash
$ styx logs list -h
Usage: styx logs list [OPTIONS]

List available logs

Global Options:
        -w, --watch             Display and update informations about logs
        -f, --format string     Output format [text|json] (default "text")
        -H, --host string       Server to connect to (default "http://localhost:8000")
        -h, --help              Display help
```

### Example

```bash
$ styx logs list
NAME            STATUS          RECORD COUNT            FILE SIZE               START POSITION          END POSITION
myLog           ok              38                      557                     0                       38
test            ok              3000075                 1524000942              0                       3000075
```

## Create log

### Usage

```bash
$ styx logs create -h
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
        -H, --host string               Server to connect to (default "http://localhost:8000")
        -h, --help                      Display help
```

### Example

```bash
$ styx logs create myLog
name:                   myLog
status:                 ok
record_count:           0
file_size:              0
start_position:         0
end_position:           0
```

## Get log

### Usage

```bash
$ styx logs get -h
Usage: styx logs get NAME [OPTIONS]

Show log details

Global Options:
        -f, --format string     Output format [text|json] (default "text")
        -H, --host string       Server to connect to (default "http://localhost:8000")
        -h, --help              Display help
```

### Example

```bash
$ styx logs get myLog
name:                   myLog
status:                 ok
record_count:           38
file_size:              557
start_position:         0
end_position:           38
```

## Delete log

### Usage

```bash
$ styx logs delete -h
Usage: styx logs delete NAME [OPTIONS]

Delete a log

Global Options:
        -H, --host string       Server to connect to (default "http://localhost:8000")
        -h, --help              Display help
```

### Example

```bash
$ styx logs delete myLog
```

## Backup log

### Usage

```bash
$ styx logs backup -h
Usage: styx logs backup NAME [OPTIONS]

Backup log

Global Options:
        -H, --host string       Server to connect to (default "http://localhost:8000")
        -h, --help              Display help
```

### Example

```bash
$ styx logs backup myLog >> myLog.backup.tar.gz
```

## Restore log

### Usage

```bash
$ styx logs restore -h
Usage: styx logs restore NAME [OPTIONS]

Restore log

Global Options:
        -H, --host string       Server to connect to (default "http://localhost:8000")
        -h, --help              Display help
```

### Example

```bash
$ styx logs restore restoredLog < myLog.backup.tar.gz
```

## Write to a log

### Usage

```bash
$ styx logs write -h
Usage: styx logs write NAME [OPTIONS]

Write to log, input is expected to be line delimited record payloads

Options:
        -u, --unbuffered        Do not buffer writes
        -b, --binary            Process input as binary records
        -l, --line-ending       Line end [cr|lf|crlf] for non binary record output

Global Options:
        -H, --host string       Server to connect to (default "http://localhost:8000")
        -h, --help              Display help
```

### Example

```bash
$ styx logs write myLog
>my first record
>my second record
```

## Read from a log

### Usage

```bash
$ styx logs read -h
Usage: styx logs read NAME [OPTIONS]

Read from log and output line delimited record payloads

Options:
        -P, --position int      Position to start reading from (default 0)
        -w, --whence string     Reference from which position is computed [origin|start|end] (default "start")
        -n, --count int         Maximum count of records to read (cannot be used in association with --follow)
        -F, --follow            Wait for new records when reaching end of stream
        -u, --unbuffered        Do not buffer read
        -b, --binary            Output binary records
        -l, --line-ending       Line end [cr|lf|crlf] for non binary record output

Global Options:
        -H, --host string       Server to connect to (default "http://localhost:8000")
        -h, --help              Display help
```

### Example

```bash
$ styx logs read myLog
my first record
my second record
```