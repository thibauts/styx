package main

import (
	"encoding/json"
	logger "log"
	"time"

	styx "github.com/dataptive/styx/pkg/client"
	"github.com/dataptive/styx/pkg/log"
)

type Event struct {
	Timestamp int64  `json:"timestamp"`
	Payload   string `json:"payload"`
}

func main() {
	client := styx.NewClient("http://localhost:7123")

	producer, err := client.NewProducer("fast", styx.DefaultProducerOptions)
	if err != nil {
		logger.Fatal(err)
	}
	defer producer.Close()

	count := 0
	for {
		event := Event{
			Timestamp: time.Now().Unix(),
			Payload:   "Hello, Styx !",
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

		if count%1000 == 0 {
			logger.Printf("produced %d records", count)
		}
	}

	producer.Flush()
}
