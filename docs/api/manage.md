Manage logs
-----------

## Create log

Create a new log.

**POST** `/logs`

### Params

| Param                 | In    | Description                                                           | Default       |
|---------------------  |------ |---------------------------------------------------------------------  |-------------- |
| `name`  _required_    | form  | The log name.                                                         |               |
| `max_record_size`     | form  | Max record size.                                                      | `1048576`     |
| `index_after_size`    | form  | Allow creating an index entry every index_after_size bytes written.   | `1048576`     |
| `segment_max_count`   | form  | Max number of records in a segment.                                   | `-1`          |
| `segment_max_size`    | form  | Max size of a segment in bytes.                                       | `1073741824`  |
| `segment_max_age`     | form  | Max age of a segment in seconds.                                      | `-1`          |
| `log_max_count`       | form  | Max number of records in a log.                                       | `-1`          |
| `log_max_size`        | form  | Max size of a log in bytes.                                           | `-1`          |
| `log_max_age`         | form  | Max age of a log in seconds.                                          | `-1`          |

### Code samples

**Bash**

```bash
$ curl -X POST 'http://localhost:8000/logs' -d name=myLog
```

### Response

```
Status: 200 OK
```
```json
{
  "name": "myLog",
  "status": "ok",
  "record_count": 0,
  "file_size": 0,
  "start_position": 0,
  "end_position": 0
}
```

## List logs

Retrieves the details of all Styx logs.

**GET** `/logs`

### Code samples

**Bash**

```bash
$ curl -X GET 'http://localhost:8000/logs'
```

### Response

```
Status: 200 OK
```
```json
[
  {
    "name": "myLog",
    "status": "ok",
    "record_count": 1345,
    "file_size": 1845,
    "start_position": 500,
    "end_position": 845
  },
  {
    "name": "myOtherLog",
    "status": "ok",
    "record_count": 542,
    "file_size": 730,
    "start_position": 0,
    "end_position": 542
  },
]
```

## Get log by name

Retrieves the details of a log.

**GET** `/logs/{name}`

### Params 

| Name        | In      | Description                                                     | Default   |
|------------ |-------  |---------------------------------------------------------------- |---------- |
| `name`      | path    | Log name.                                                       |           |

### Code samples

**Bash**

```bash
$ curl -X GET 'http://localhost:8000/logs/myLog'
```

### Response

```
Status: 200 OK
```
```json
{
  "name": "myLog",
  "status": "ok",
  "record_count": 1345,
  "file_size": 1845,
  "start_position": 500,
  "end_position": 845
}
```

## Delete log

Permanently delete a log and its data.

**DELETE** `/logs/{name}`

### Params 

| Name        | In      | Description                                                     | Default   |
|------------ |-------  |---------------------------------------------------------------- |---------- |
| `name`      | path    | Log name.                                                       |           |

### Code samples

**Bash**

```bash
$ curl -X DELETE 'http://localhost:8000/logs/myLog'
```

## Truncate log

Empty a log of all its records.

**POST** `/logs/{name}/truncate`

### Params 

| Name        | In      | Description                                                     | Default   |
|------------ |-------  |---------------------------------------------------------------- |---------- |
| `name`      | path    | Log name.                                                       |           |

### Code samples

**Bash**

```bash
$ curl -X POST 'http://localhost:8000/logs/myLog/truncate'
```

### Response

```
Status: 200 OK
```

## Backup log

Download a backup of the log.

**GET** `/logs/{name}/backup`

### Params 

| Name        | In      | Description                                                     | Default   |
|------------ |-------  |---------------------------------------------------------------- |---------- |
| `name`      | path    | Log name.                                                       |           |

### Code samples

**Bash**

```bash
$ curl -X GET 'http://localhost:8000/logs/myLog/backup' -o myLogBackup.tar.gz
```

### Response

```
Status: 200 OK
```

Response body contains binary backup archive.

## Restore log

Imports a previously backed up log archive.

**POST** `/logs/restore`

### Params 

| Name                | In       | Description                                                     | Default   |
|-------------------- |--------- |---------------------------------------------------------------- |---------- |
| `name` _Required_   | query    | Log name.                                                       |           |
|                     | body     | Binay backup archive.                                           |           |

### Code samples

**Bash**

```bash
$ curl -X POST 'http://localhost:8000/logs/restore?name=myRestoredLog' --data-binary '@myLogBackup.tar.gz'  
```

### Response

```
Status: 200 OK
```
