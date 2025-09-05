// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Reg = prometheus.NewRegistry()
	// TO-DO this is where we would implement a custom collector
)

func Init(metricsAddr string) {
	Reg.MustRegister(prometheus.NewGoCollector())
	Reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(metricsAddr, nil)
}
