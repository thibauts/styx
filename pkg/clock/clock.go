// Copyright 2021 Dataptive SAS.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
