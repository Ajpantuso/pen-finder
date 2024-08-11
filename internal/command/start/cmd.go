// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package start

import (
	"fmt"
	"log"
	"time"

	"github.com/ajpantuso/pen-finder/internal/metrics"
	"github.com/ajpantuso/pen-finder/internal/recorder/prometheus"
	"github.com/ajpantuso/pen-finder/internal/server"
	"github.com/go-logr/zapr"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func NewCommand() *cobra.Command {
	flags := flags{
		BindAddr:        ":8080",
		MetricsBindAddr: ":8083",
		CertFile:        "server.crt",
		KeyFile:         "server.key",
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts server",
		RunE:  run(flags),
	}
	flags.AddFlags(cmd.Flags())

	return cmd
}

func run(flags flags) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
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
			server.WithBindAddr(flags.BindAddr),
			server.WithKeyFile(flags.KeyFile),
			server.WithCertFile(flags.CertFile),
			server.WithLogger{Logger: logger},
			server.WithRunTimeout(10*time.Second),
			server.WithRecorder{Recorder: recorder},
		)

		go func() {
			errCh <- srv.Serve(ctx)
		}()

		metricsSrv := metrics.NewServer(metrics.WithBindAddr(flags.MetricsBindAddr), metrics.WithGatherer{Gatherer: registry})

		go func() {
			errCh <- metricsSrv.Serve(ctx)
		}()

		for {
			select {
			case err := <-errCh:
				return fmt.Errorf("unexpected error: %w", err)
			case <-ctx.Done():
				logger.Info("received shutdown signal")

				if err := multierr.Combine(<-errCh, <-errCh); err != nil {
					return fmt.Errorf("shutting down: %w", err)
				}

				return nil
			}
		}
	}
}

type flags struct {
	BindAddr        string
	MetricsBindAddr string
	CertFile        string
	KeyFile         string
}

func (f *flags) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&f.BindAddr, "bind-addr", f.BindAddr, "Address for server to listen on")
	flags.StringVar(&f.MetricsBindAddr, "metrics-bind-addr", f.MetricsBindAddr, "Address for metrics server to listen on")
	flags.StringVar(&f.CertFile, "cert-file", f.CertFile, "Path to server TLS certificate")
	flags.StringVar(&f.KeyFile, "key-file", f.KeyFile, "Path to server TLS private key")
}
