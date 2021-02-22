// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package statsd

import (
	"net"
	"time"
	"sync"

	"gitlab.com/dataptive/styx/logger"
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
		sock: sock,
		address: address,
		resolvedDest: resolved,
		ticker: ticker,
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
