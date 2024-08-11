// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewServer(opts ...Option) *Server {
	var cfg Config

	cfg.Options(opts...)

	return &Server{
		cfg: cfg,
	}
}

type Server struct {
	cfg Config
}

func (s *Server) Serve(ctx context.Context) error {
	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.HandlerFor(s.cfg.Gatherer, promhttp.HandlerOpts{}))

	srv := http.Server{
		Addr:    s.cfg.BindAddr,
		Handler: handler,
	}

	errCh := make(chan error)

	go func() {
		errCh <- srv.ListenAndServe()
	}()

	for {
		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return srv.Shutdown(context.Background())
		}
	}
}

type Config struct {
	BindAddr string
	Gatherer prometheus.Gatherer
}

func (c *Config) Options(opts ...Option) {
	for _, opt := range opts {
		opt.ConfigureServer(c)
	}
}

type Option interface {
	ConfigureServer(*Config)
}

type WithBindAddr string

func (w WithBindAddr) ConfigureServer(c *Config) {
	c.BindAddr = string(w)
}

type WithGatherer struct {
	Gatherer prometheus.Gatherer
}

func (w WithGatherer) ConfigureServer(c *Config) {
	c.Gatherer = w.Gatherer
}
