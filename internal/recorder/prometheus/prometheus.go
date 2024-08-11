// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"github.com/ajpantuso/pen-finder/internal/recorder"
	"github.com/prometheus/client_golang/prometheus"
)

func NewRecorder(opts ...Option) (*Recorder, error) {
	var cfg Config

	cfg.Options(opts...)

	matchesOpts := prometheus.GaugeOpts{
		Name: "matches",
	}
	matches := prometheus.NewGaugeVec(matchesOpts, []string{"source", "name", "url"})

	if err := cfg.Registerer.Register(matches); err != nil {
		return nil, err
	}

	return &Recorder{
		matches: matches,
	}, nil
}

type Recorder struct {
	matches *prometheus.GaugeVec
}

func (r *Recorder) RecordProduct(product recorder.Product) error {
	gauge, err := r.matches.GetMetricWithLabelValues(product.Source, product.Name, product.URL)
	if err != nil {
		return err
	}

	gauge.Inc()

	return nil
}

type Config struct {
	Registerer prometheus.Registerer
}

func (c *Config) Options(opts ...Option) {
	for _, opt := range opts {
		opt.ConfigureRecorder(c)
	}
}

func (c *Config) Default() {
	if c.Registerer == nil {
		c.Registerer = prometheus.NewRegistry()
	}
}

type Option interface {
	ConfigureRecorder(*Config)
}

type WithRegisterer struct {
	Registerer prometheus.Registerer
}

func (w WithRegisterer) ConfigureRecorder(c *Config) {
	c.Registerer = w.Registerer
}
