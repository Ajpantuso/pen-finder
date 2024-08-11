// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package scraper

import (
	"context"
	"sync"

	"go.uber.org/multierr"
)

type Runner interface {
	Run(context.Context, ...RunOption) error
}

type RunConfig struct {
	Scrapers      []Scraper
	ScrapeOptions []ScrapeOption
}

func (c *RunConfig) Options(opts ...RunOption) {
	for _, opt := range opts {
		opt.ConfigureRun(c)
	}
}

type RunOption interface {
	ConfigureRun(*RunConfig)
}

type WithScrapers []Scraper

func (w WithScrapers) ConfigureRun(c *RunConfig) {
	c.Scrapers = append(c.Scrapers, w...)
}

type WithScrapeOptions []ScrapeOption

func (w WithScrapeOptions) ConfigureRun(c *RunConfig) {
	c.ScrapeOptions = append(c.ScrapeOptions, w...)
}

func NewParallelRunner() *ParallelRunner {
	return &ParallelRunner{}
}

type ParallelRunner struct{}

func (r *ParallelRunner) Run(ctx context.Context, opts ...RunOption) error {
	var cfg RunConfig

	cfg.Options(opts...)

	var wg sync.WaitGroup
	errCh := make(chan error)

	for _, scraper := range cfg.Scrapers {
		scraper := scraper
		wg.Add(1)

		go func() {
			errCh <- scraper.Scrape(ctx, cfg.ScrapeOptions...)

			wg.Done()
		}()
	}

	go func() {
		wg.Wait()

		close(errCh)
	}()

	var finalErr error

	for err := range errCh {
		multierr.AppendInto(&finalErr, err)
	}

	return finalErr
}
