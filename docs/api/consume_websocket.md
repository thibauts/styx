Consume with WebSocket
----------------------

Consume records using WebSocket protocol.

**GET** `/logs/{name}/records`

Upgrade: websocket  
Connection: Upgrade  

### Params 

| Name       	| In    	| Description                                                    	    | Default  	|
|------------	|-------	|-------------------------------------------------------------------	|----------	|
| `name`     	| path  	| Log name.                                                      	    |          	|
| `whence`   	| query 	| Allowed values are `origin`, `start` and `end`.                	    | `origin` 	|
| `position` 	| query 	| Whence relative position from which the records are consumed from. 	| `0`      	|

### Response 

```
Status: 101 Switching protocol
```

### Code samples

**Wsdump** (_Requires [websocket-client](https://pypi.org/project/websocket-client-py3/) package._)
```bash
$ wsdump.py ws://localhost:7123/logs/myLog/records
```

**Python** (_Requires [websocket-client](https://pypi.org/project/websocket-client-py3/) package._)

```python
import websocket

ws = websocket.create_connection('ws://localhost:7123/logs/myLog/records')
while True:
  record = ws.recv()
  print(record)
```

**Go** (_Requires [github.com/gorilla/websocket](http://github.com/gorilla/websocket) package._)

```golang
import (
  "fmt"
  "log"
  "github.com/gorilla/websocket"
)

dialer := websocket.Dialer{}

conn, res, err := dialer.Dial("ws://localhost:7123/logs/myLog/records?whence=start&position=0", nil)
if err != nil {
  log.Fatal(err)
}

for {
    _, record, err := conn.ReadMessage()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(record))
}
```
