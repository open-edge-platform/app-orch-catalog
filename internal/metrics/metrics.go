// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"errors"
	"net/http"

	"github.com/open-edge-platform/orch-library/go/dazl"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var log = dazl.GetPackageLogger()

var (
	Reg = prometheus.NewRegistry()
	// TODO: this is where we would implement a custom collector
)

func Init(metricsAddr string) {
	Reg.MustRegister(prometheus.NewGoCollector())
	Reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(Reg, promhttp.HandlerOpts{}))
	go func() {
		if err := http.ListenAndServe(metricsAddr, mux); !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("Failed to start metrics server on %s: %v", metricsAddr, err)
		}
	}()
}
