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

package prometheus

import (
	"github.com/dataptive/styx/pkg/log"

	prom "github.com/prometheus/client_golang/prometheus"
)

type PrometheusReporter struct {
	logRecordCount *prom.GaugeVec
	logFileSize    *prom.GaugeVec
}

func NewPrometheusReporter() (pp *PrometheusReporter) {

	logRecordCount := prom.NewGaugeVec(
		prom.GaugeOpts{
			Name: "log_record_count",
			Help: "Current record count",
		},
		[]string{"log"},
	)

	logFileSize := prom.NewGaugeVec(
		prom.GaugeOpts{
			Name: "log_file_size",
			Help: "Current log file size",
		},
		[]string{"log"},
	)

	prom.MustRegister(logRecordCount)
	prom.MustRegister(logFileSize)

	pp = &PrometheusReporter{
		logRecordCount: logRecordCount,
		logFileSize:    logFileSize,
	}

	return pp
}

func (pp *PrometheusReporter) Close() (err error) {

	return nil
}

func (pp *PrometheusReporter) ReportLogStats(name string, stats log.Stat) (err error) {

	recordCount := float64(stats.EndPosition - stats.StartPosition)
	pp.logRecordCount.
		With(prom.Labels{"log": name}).
		Set(recordCount)

	fileSize := float64(stats.EndOffset - stats.StartOffset)
	pp.logFileSize.
		With(prom.Labels{"log": name}).
		Set(fileSize)

	return nil
}
