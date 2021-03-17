// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package benchmark

import (
	"fmt"
	"io"
	"time"

	"github.com/dataptive/styx/logger"

	"github.com/dataptive/styx/client"
	"github.com/dataptive/styx/cmd"
	"github.com/dataptive/styx/log"

	"github.com/spf13/pflag"
)

const benchmarkRunUsage = `
Usage: styx benchmark [OPTIONS]

Run benchmarks

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

const benchmarkLogo = `             __
     _______/  |_ ___.__.___  ___
    /  ___/\   __<   |  |\  \/  /
    \___ \  |  |  \___  | >    <
   /____  > |__|  / ____|/__/\_ \  BENCHMARK
        \/        \/           \/
`

func displayMetrics(prefix string, producedRecords int, producedBytes int, elapsed time.Duration) {

	recordRate := float64(producedRecords) / float64(elapsed.Seconds())
	byteRate := float64(producedBytes) / float64(elapsed.Seconds())
	elapsedSeconds := float64(elapsed.Seconds())

	fmt.Printf(
		"  %s %d records in %.2f seconds (%.2f records/s, %.2f MB/s)",
		prefix,
		producedRecords,
		elapsedSeconds,
		recordRate,
		byteRate / float64(1 << 20),
	)
}

func benchmarkProduce(host string, name string, size int, count int) (err error) {

	fmt.Printf("--------------------------------------------------------------------------------\n")
	fmt.Printf("* benchmarking PRODUCE with %d records of size %d\n", count, size)

	c := client.NewClient(host)

	_, err = c.CreateLog(name, client.DefaultLogConfig)
	if err != nil {
		return err
	}
	defer c.DeleteLog(name)

	producer, err := c.NewProducer(name, client.DefaultProducerOptions)
	if err != nil {
		return err
	}
	defer producer.Close()

	payload := make([]byte, size)
	r := log.Record(payload)

	producedRecords := 0
	producedBytes := 0
	start := time.Now()

	fmt.Printf("  starting ...\n")

	for {
		n, err := producer.Write(&r)
		if err != nil {
			return err
		}

		producedRecords += 1
		producedBytes += n

		if producedRecords % 10000 == 0 {
			elapsed := time.Since(start)
			fmt.Printf("\033[1K\r")
			displayMetrics("produced", producedRecords, producedBytes, elapsed)
		}

		if producedRecords == count {
			break
		}
	}

	producer.Flush()

	fmt.Printf("\n")
	fmt.Printf("  done\n")

	return nil
}

func benchmarkConsume(host string, name string, size int, count int) (err error) {

	fmt.Printf("--------------------------------------------------------------------------------\n")
	fmt.Printf("* benchmarking CONSUME with %d records of size %d\n", count, size)
	fmt.Printf("  preparing log ...\n")

	c := client.NewClient(host)

	_, err = c.CreateLog(name, client.DefaultLogConfig)
	if err != nil {
		return err
	}
	defer c.DeleteLog(name)

	producer, err := c.NewProducer(name, client.DefaultProducerOptions)
	if err != nil {
		return err
	}

	payload := make([]byte, size)
	r := log.Record(payload)

	for {
		_, err := producer.Write(&r)
		if err != nil {
			return err
		}

		count -= 1

		if count == 0 {
			break
		}
	}

	producer.Flush()
	producer.Close()

	fmt.Printf("  starting ...\n")

	consumer, err := c.NewConsumer(name, client.DefaultConsumerParams, client.DefaultConsumerOptions)
	if err != nil {
		return err
	}
	defer consumer.Close()

	r = log.Record{}

	consumedRecords := 0
	consumedBytes := 0
	start := time.Now()

	for {
		n, err := consumer.Read(&r)

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		consumedRecords += 1
		consumedBytes += n

		if consumedRecords % 10000 == 0 {
			elapsed := time.Since(start)
			fmt.Printf("\033[1K\r")
			displayMetrics("consumed", consumedRecords, consumedBytes, elapsed)
		}
	}

	fmt.Printf("\n")
	fmt.Printf("  done\n")

	return nil
}

func RunBenchmark(args []string) {

	logger.SetLevel(logger.LevelInfo)

	logName := "benchmark"

	runOpts := pflag.NewFlagSet("benchmark", pflag.ContinueOnError)
	host := runOpts.StringP("host", "H", "http://localhost:7123", "")
	isHelp := runOpts.BoolP("help", "h", false, "")
	runOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, benchmarkRunUsage)
	}

	err := runOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, benchmarkRunUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, benchmarkRunUsage)
	}

	fmt.Printf("%s\n", benchmarkLogo)

	params := [][]int{
		{10, 10000000},
		{100, 1000000},
		{1000, 100000},
	}

	for _, param := range params {
		err = benchmarkProduce(*host, logName, param[0], param[1])
		if err != nil {
			cmd.DisplayError(err)
		}
	}

	for _, param := range params {
		err = benchmarkConsume(*host, logName, param[0], param[1])
		if err != nil {
			cmd.DisplayError(err)
		}
	}

}
