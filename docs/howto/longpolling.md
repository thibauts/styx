HTTP long polling
----------------------

Styx HTTP api provides means to read log records using long polling.

Assuming you have never consumed any records from `myLog` and you want to retrieve its content in batchs of 100 line delimited records.

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import request

res = requests.get(
  'http://localhost:8000/logs/myLog/records',
  headers={
    'Accept': 'application/vnd.styx.line-delimited;line-ending=lf',
  },
  params={
    'count': '100'
  }
)
```

**Go**

```golang
import (
  "log"
  "net/http"
)

client := &http.Client{}

req, err := http.NewRequest(http.MethodGet, "http://localhost:8000/logs/myLog/records?count=100", nil)
if err != nil {
  log.Fatal(err)
}

req.Header.Add("Accept", "application/vnd.styx.line-delimited;line-ending=lf")

res, err := client.Do(req)
if err != nil {
  log.Fatal(err)
}
```

Assuming you effectively read 100 records, you have to increment the `position` query param with the number of processed records to consume the next ones.

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import request

res = requests.get(
  'http://localhost:8000/logs/myLog/records',
  headers={
    'Accept': 'application/vnd.styx.line-delimited;line-ending=lf',
  },
  params={
    'count': '100',
    'position': '100',
  }
)
```

**Go**

```golang
import (
  "log"
  "net/http"
)

client := &http.Client{}

req, err := http.NewRequest(http.MethodGet, "http://localhost:8000/logs/myLog/records?position=100&count=100", nil)
if err != nil {
  log.Fatal(err)
}

req.Header.Add("Accept", "application/vnd.styx.line-delimited;line-ending=lf")

res, err := client.Do(req)
if err != nil {
  log.Fatal(err)
}
```

If there was only 50 records to read, a response will be returned with the remaining records.
Now you want to wait for new records to be written by adding the `follow` query param.

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import request

res = requests.get(
  'http://localhost:8000/logs/myLog/records',
  headers={
    'Accept': 'application/vnd.styx.line-delimited;line-ending=lf',
    'X-Styx-Timeout': '30'
  },
  params={
    'count': '100',
    'position': '100',
    'follow': 'true',
  }
)
```

**Go**

```golang
import (
  "log"
  "net/http"
)

client := &http.Client{}

req, err := http.NewRequest(http.MethodGet, "http://localhost:8000/logs/myLog/records?position=150&count=100&follow=true", nil)
if err != nil {
  log.Fatal(err)
}

req.Header.Add("Accept", "application/vnd.styx.line-delimited;line-ending=lf")
req.Header.Add("X-Styx-Timeout", "30")

res, err := client.Do(req)
if err != nil {
  log.Fatal(err)
}
```

The request will hang, up to the number of seconds specified by `X-Styx-Timeout` header, or until new records are written to the log.
As soon as new records are available these will be returned in the response, closing the HTTP communication.

With an infinite loop you can easily create a long running consummer using long polling.   
Here is the full code example.

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
import requests

position = 0
while True:
  res = requests.get(
    'http://localhost:8000/logs/myLog/records',
    headers={
      'Accept': 'application/vnd.styx.line-delimited;line-ending=lf',
      'X-Styx-Timeout': '30'
    },
    params={
      'count': '100',
      'position': str(position),
      'follow': 'true',
    }
  )

  for line in res.iter_lines():
    print(line.decode('utf-8'))      
    position += 1
```

**Go**

```golang
import (
  "io"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "string"
)

client := &http.Client{}

position := int64(0)

for {
  url := "http://localhost:8000/logs/myLog/records?position=" + strconv.FormatInt(position, 10) + "&count=100&follow=true"

  req, err := http.NewRequest(http.MethodGet, url, nil)
  if err != nil {
    log.Fatal(err)
  }

  req.Header.Set("Accept", "application/vnd.styx.line-delimited;line-ending=lf")
  req.Header.Set("X-Styx-Timeout", "30")

  res, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }

  if res.StatusCode != http.StatusOK {
    continue
  }

  buf, err := ioutil.ReadAll(res.Body)
  if err != nil {
    log.Fatal(err)
  }

  records := string.Split(string(buf), "\n")

  position += int64(len(records))

  io.Copy(os.Stdout, buf)

  res.Body.Close()
}
```