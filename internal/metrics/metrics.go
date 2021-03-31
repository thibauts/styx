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

package metrics

import (
	"github.com/dataptive/styx/pkg/log"

	"github.com/dataptive/styx/internal/metrics/prometheus"
	"github.com/dataptive/styx/internal/metrics/statsd"
)

type Reporter interface {
	ReportLogStats(string, log.Stat) error

	Close() error
}

type MetricsReporter struct {
	reporters []Reporter
}

func NewMetricsReporter(config Config) (mp *MetricsReporter, err error) {

	var reporters []Reporter

	pp := prometheus.NewPrometheusReporter()
	reporters = append(reporters, pp)

	if config.Statsd != nil {
		sp, err := statsd.NewStatsdReporter(*config.Statsd)
		if err != nil {
			return nil, err
		}

		reporters = append(reporters, sp)
	}

	mp = &MetricsReporter{
		reporters: reporters,
	}

	return mp, nil
}

func (mp *MetricsReporter) ReportLogStats(name string, stats log.Stat) (err error) {

	for _, reporter := range mp.reporters {
		reporter.ReportLogStats(name, stats)
	}

	return nil
}

func (mp *MetricsReporter) Close() (err error) {

	for _, reporter := range mp.reporters {
		reporter.Close()
	}

	return nil
}
