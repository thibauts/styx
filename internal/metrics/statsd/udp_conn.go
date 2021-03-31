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

package statsd

import (
	"net"
	"sync"
	"time"

	"github.com/dataptive/styx/pkg/logger"
)

var resolvePeriod = 10 * time.Minute

type UDPConn struct {
	sock         *net.UDPConn
	address      string
	resolvedDest *net.UDPAddr
	ticker       *time.Ticker
	resolveLock  sync.RWMutex
}

func NewUDPConn(address string) (uc *UDPConn, err error) {

	// Listen on all available IPs,
	// on an automatically chosen port.
	sock, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return nil, err
	}

	resolved, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(resolvePeriod)

	uc = &UDPConn{
		sock:         sock,
		address:      address,
		resolvedDest: resolved,
		ticker:       ticker,
	}

	go uc.resolver()

	return uc, nil
}

func (uc *UDPConn) Close() (err error) {

	uc.ticker.Stop()

	err = uc.sock.Close()
	if err != nil {
		return err
	}

	return nil
}

func (uc *UDPConn) Write(p []byte) (n int, err error) {

	uc.resolveLock.RLock()
	defer uc.resolveLock.RUnlock()

	n, err = uc.sock.WriteTo(p, uc.resolvedDest)
	if err != nil {
		return n, err
	}

	return n, nil
}

// resolver periodically resolve given address
func (uc *UDPConn) resolver() {

	for range uc.ticker.C {

		// When the provided address contains an host name
		// and host name is resolved to multiple IPs
		// one IP is arbitrarily chosen.
		resolved, err := net.ResolveUDPAddr("udp", uc.address)
		if err != nil {
			logger.Warn("statsd: unable to resolve address")
			continue
		}

		uc.resolveLock.Lock()
		uc.resolvedDest = resolved
		uc.resolveLock.Unlock()
	}
}
