Read with Styx protocol
------------------------

Read records using [Styx protocol](/docs/api/styx_protocol.md).

**GET** `/logs/{name}/records`  

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

params := client.DefaultConsumerParams
// params.Follow = true

consumer, err := c.NewConsumer("test", params, client.DefaultConsumerOptions)
if err != nil {
	logger.Fatal(err)
}
defer consumer.Close()

r := log.Record{}

for {
	_, err := consumer.Read(&r)
	if err == io.EOF {
		break
	}

	if err != nil {
		logger.Fatal(err)
	}

	logger.Println(string(r))
}
```
