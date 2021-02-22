Write with HTTP
---------------

Write records using HTTP protocol.

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

#### Write a record

**Curl**

```bash
$ curl -X POST 'http://localhost:8000/logs/myLog/records' -d 'my record content'
```

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import requests

requests.post('http://localhost:8000/logs/myLog/records', data=b'my record content')
```

**Go**

```golang
import (
  "bytes"
  "net/http"
)

client := &http.Client{}
client.Post(
  "http://localhost:8000/logs/myLog/records", 
  "application/octet-stream", 
  bytes.NewReader([]byte("my record content")),
)
```

#### Write line delimited records

**Curl**

```bash
$ curl -X POST 'http://localhost:8000/logs/myLog/records' \
  -H 'Content-Type: application/vnd.styx.line-delimited;line-ending=lf' \
  -d $'my record content\nmy record content\n'
```

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import requests

requests.post(
  'http://localhost:8000/logs/myLog/records',
  headers={
    'Content-Type': 'application/vnd.styx.line-delimited;line-ending=lf'
  },
  data=b''.join([b'my record content\n' for i in range(10)])
)
```

**Go**

```golang
import (
  "bytes"
  "net/http"
  "string"
)

records := strings.Repeat("my record content\n", 10)

client := &http.Client{}
client.Post(
  "http://localhost:8000/logs/myLog/records",
  "application/vnd.styx.line-delimited;line-ending=lf", 
  bytes.NewReader([]byte(records)),
)
```
