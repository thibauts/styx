Write with Styx protocol
------------------------

Write records using [Styx protocol](/docs/api/styx_protocol.md).

**POST** `/logs/{name}/records`  

Upgrade: styx/0  
Connection: Upgrade  

### Params 

| Name             	| In     	| Description                                                                                         	| Default 	|
|------------------	|--------	|-----------------------------------------------------------------------------------------------------	|---------	|
| `name`           	| path   	| Log name.                                                                                           	|         	|
| `X-Styx-Timeout` 	| header 	| The maximum amount of seconds the peer will keep the connection opened whithout receiving messages. 	|         	|

### Response 

```
Status: 101 Switching protocol
```

### Code samples

**Go** (_Requires [styx/client](), [styx/log]() packages._)

```golang
c := client.NewClient("http://localhost:8000")

producer, err := c.NewProducer("test", client.DefaultProducerOptions)
if err != nil {
	logger.Fatal(err)
}
defer producer.Close()

r := log.Record([]byte("Hello, Styx !"))

for i := 0; i < 10; i++ {
	_, err := producer.Write(&r)
	if err != nil {
		logger.Fatal(err)
	}
}

producer.Flush()
```
