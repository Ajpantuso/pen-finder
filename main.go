// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ajpantuso/pen-finder/internal/metrics"
	"github.com/ajpantuso/pen-finder/internal/recorder/prometheus"
	"github.com/ajpantuso/pen-finder/internal/server"
	"github.com/go-logr/zapr"
	prom "github.com/prometheus/client_golang/prometheus"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	errCh := make(chan error, 2)

	registry := prom.NewRegistry()

	recorder, err := prometheus.NewRecorder(prometheus.WithRegisterer{Registerer: registry})
	if err != nil {
		log.Fatal(err)
	}

	zlog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	logger := zapr.NewLogger(zlog)
	srv := server.NewDefaultServer(
		server.WithBindAddr(":8080"),
		server.WithKeyFile("server.key"),
		server.WithCertFile("server.cert"),
		server.WithLogger{Logger: logger},
		server.WithRunTimeout(10*time.Second),
		server.WithRecorder{Recorder: recorder},
	)

	go func() {
		errCh <- srv.Serve(ctx)
	}()

	metricsSrv := metrics.NewServer(metrics.WithBindAddr(":8083"), metrics.WithGatherer{Gatherer: registry})

	go func() {
		errCh <- metricsSrv.Serve(ctx)
	}()

	for {
		select {
		case err := <-errCh:
			logger.Error(err, "unexpected error")

			os.Exit(1)
		case <-ctx.Done():
			logger.Info("received shutdown signal")

			if err := multierr.Combine(<-errCh, <-errCh); err != nil {
				logger.Error(err, "shutting down")
			}

			return
		}
	}
}
