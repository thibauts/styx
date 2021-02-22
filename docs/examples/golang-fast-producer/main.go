package main

import (
	"encoding/json"
	logger "log"
	"time"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/log"
)

type Event struct {
	Timestamp int64 `json:"timestamp"`
	Payload string `json:"payload"`
}

func main() {
	c := client.NewClient("http://localhost:8000")

	producer, err := c.NewProducer("fast", client.DefaultProducerOptions)
	if err != nil {
		logger.Fatal(err)
	}
	defer producer.Close()

	count := 0
	for {
		event := Event{
			Timestamp: time.Now().Unix(),
			Payload: "Hello, Styx !",
		}

		payload, err := json.Marshal(event)
		if err != nil {
			logger.Fatal(err)
		}

		r := log.Record(payload)

		_, err = producer.Write(&r)
		if err != nil {
			logger.Fatal(err)
		}

		count++

		if count % 1000 == 0 {
			logger.Printf("sent %d records", count)
		}
	}

	producer.Flush()
}
