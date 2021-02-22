Read using HTTP
---------------

Read records using HTTP protocol.

**GET** `/logs/{name}/records`  

### Params 

| Name             	| In     	| Description                                                                                                                  	| Default                    	|
|------------------	|--------	|------------------------------------------------------------------------------------------------------------------------------	|----------------------------	|
| `name`           	| path   	| Log name.                                                                                                                    	|                            	|
| `whence`         	| query  	| Allowed values are `origin`, `start` and `end`.                                                                              	| `origin`                   	|
| `position`       	| query  	| Whence relative position from which the records are read from.                                                               	| `0`                        	|
| `count`          	| query  	| Limits the number of records to read, `-1` means no limitation.<br>Not available with `application/octet-stream` media type. 	| `-1`                       	|
| `follow`         	| query  	| Read will block until new records are written to the log.<br>Not available with `application/octet-stream` media type.       	| `false`                    	|
| `Accept`         	| header 	| See [Media-Types](/docs/api/media_types.md) for allowed values.                                                              	| `application/octet-stream` 	|
| `X-Styx-Timeout` 	| header 	| Number of seconds before timing out when waiting for new records with the `follow` query param.                              	|                            	|

### Response 

```
Status: 200 OK
```

Response contains records formatted according to `Accept`header.  

### Codes samples

#### Read the first available record

**Curl**

```bash
$ curl -X GET 'http://localhost:8000/logs/myLog/records&whence=start'
```

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import requests

res = requests.get(
  'http://localhost:8000/logs/myLog/records',
  params={
    'whence': 'start'
  }
)

print(res.text)
```

**Go**

```golang
import (
  "io/ioutil"
  "fmt"
  "log"
  "net/http"
)

client := &http.Client{}

res, err := client.Get("http://localhost:8000/logs/myLog/records?whence=start")
if err != nil {
  log.Fatal(err)
}

record, err := ioutil.ReadAll(res.Body)
if err != nil {
  log.Fatal(err)
}

fmt.Println(record)
```

#### Read first ten available records.

**Curl**

```bash
$ curl -X GET 'http://localhost:8000/logs/myLog/records?whence=start&count=10' \
  -H 'Accept: application/vnd.styx.line-delimited;line-ending=lf'
```

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import requests

res = requests.get(
  'http://localhost:8000/logs/myLog/records',
  params={
    'whence': 'start',
    'count': 10
  },
  headers={
    'Accept': 'application/vnd.styx.line-delimited;line-ending=lf'
  }
)

for record in res.text.iter_lines():
  print(record)
```

**Go**

```golang
import (
  "io"
  "log"
  "net/http"  
  "os"
)

client := &http.Client{}

req, err := http.NewRequest(http.MethodGet, "http://localhost:8000/logs/myLog/records?whence=start&count=10", nil)
if err != nil {
  log.Fatal(err)
}

req.Header.Add("Accept", "application/vnd.styx.line-delimited;line-ending=lf")

res, err := client.Do(req)
if err != nil {
  log.Fatal(err)
}

io.Copy(os.Stdout, res.Body)
```