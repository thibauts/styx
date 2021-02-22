// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package clock

import (
	"time"
	"sync/atomic"
)

type Clock struct {
	updaterStop chan struct{}
	updaterDone chan struct{}
	ticker *time.Ticker
	nanoTimestamp int64
}

func New(resolution time.Duration) (c *Clock) {

	ticker := time.NewTicker(resolution)

	c = &Clock{
		updaterStop: make(chan struct {}),
		updaterDone: make(chan struct {}),
		ticker: ticker,
		nanoTimestamp: time.Now().UnixNano(),
	}

	go c.updater()

	go func() {
		// Realign ticker to whole ticks.
		now := time.Now()
		drift := now.Sub(now.Truncate(resolution))
		time.Sleep(resolution - drift)
		ticker.Reset(resolution)

		// Update timestamp to compensate for delayed
		// update introduced by ticker reset.
		atomic.StoreInt64(&c.nanoTimestamp, time.Now().UnixNano())
	}()

	return c
}

func (c *Clock) Stop() {

	c.ticker.Stop()
	c.updaterStop <- struct{}{}
	<- c.updaterDone
}

func (c *Clock) Time() (t time.Time) {

	nanoTimestamp := atomic.LoadInt64(&c.nanoTimestamp)

	t = time.Unix(0, nanoTimestamp)

	return t
}

func (c *Clock) Unix() (timestamp int64) {

	nanoTimestamp := atomic.LoadInt64(&c.nanoTimestamp)

	return nanoTimestamp / 1e9
}

func (c *Clock) UnixNano() (nanoTimestamp int64) {

	nanoTimestamp = atomic.LoadInt64(&c.nanoTimestamp)

	return nanoTimestamp
}

func (c *Clock) updater() {

	for {
		select {
		case <- c.updaterStop:
			c.updaterDone <- struct{}{}
			return
		case t := <- c.ticker.C:
			atomic.StoreInt64(&c.nanoTimestamp, t.UnixNano())
		}
	}
}
