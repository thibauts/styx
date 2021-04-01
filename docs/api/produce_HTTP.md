Produce with HTTP
-----------------

Produce records using HTTP protocol.

**POST** `/logs/{name}/records`  

### Params

| Name           	| In     	| Description                                                     	| Default                    	|
|----------------	|--------	|-----------------------------------------------------------------	|----------------------------	|
| `name`         	| path   	| Log name.                                                       	|                            	|
| `Content-Type` 	| header 	| See [Media-Types](/docs/api/media_types.md) for allowed values. 	| `application/octet-stream` 	|

### Response 

```
Status: 200 OK
```
```json
{
  "position": 20,
  "count": 10,
}
```

### Codes samples

#### Produce a record

**Curl**

```bash
$ curl -X POST 'http://localhost:7123/logs/myLog/records' -d 'my record payload'
```

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import requests

requests.post('http://localhost:7123/logs/myLog/records', data=b'my record payload')
```

**Go**

```golang
import (
  "bytes"
  "net/http"
)

client := &http.Client{}
client.Post(
  "http://localhost:7123/logs/myLog/records", 
  "application/octet-stream", 
  bytes.NewReader([]byte("my record payload")),
)
```

#### Produce line delimited records

**Curl**

```bash
$ curl -X POST 'http://localhost:7123/logs/myLog/records' \
  -H 'Content-Type: application/vnd.styx.line-delimited;line-ending=lf' \
  -d $'my record payload\nmy record payload\n'
```

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import requests

requests.post(
  'http://localhost:7123/logs/myLog/records',
  headers={
    'Content-Type': 'application/vnd.styx.line-delimited;line-ending=lf'
  },
  data=b''.join([b'my record payload\n' for i in range(10)])
)
```

**Go**

```golang
import (
  "bytes"
  "net/http"
  "string"
)

records := strings.Repeat("my record payload\n", 10)

client := &http.Client{}
client.Post(
  "http://localhost:7123/logs/myLog/records",
  "application/vnd.styx.line-delimited;line-ending=lf", 
  bytes.NewReader([]byte(records)),
)
```
