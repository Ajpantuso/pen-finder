// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package scraper

import (
	"context"
	"errors"
	"fmt"
	"github.com/ajpantuso/pen-finder/internal/recorder"
	"regexp"

	"github.com/gocolly/colly/v2"
	"go.uber.org/multierr"
)

type Scraper interface {
	Scrape(context.Context, ...ScrapeOption) error
}

type ScrapeConfig struct {
	Recorder recorder.Recorder
}

func (c *ScrapeConfig) Options(opts ...ScrapeOption) {
	for _, opt := range opts {
		opt.ConfigureScrape(c)
	}
}

func (c *ScrapeConfig) Default() {
	if c.Recorder == nil {
		c.Recorder = recorder.NewDebugRecorder()
	}
}

type ScrapeOption interface {
	ConfigureScrape(c *ScrapeConfig)
}

func NewSimpleScraper(opts ...SimpleScraperOption) *SimpleScraper {
	var cfg SimpleScraperConfig

	cfg.Options(opts...)

	return &SimpleScraper{
		collector: colly.NewCollector(colly.Async(), colly.URLFilters(cfg.Filters...)),
		cfg:       cfg,
	}
}

type SimpleScraper struct {
	collector *colly.Collector
	cfg       SimpleScraperConfig
}

func (s *SimpleScraper) Scrape(ctx context.Context, opts ...ScrapeOption) error {
	var cfg ScrapeConfig

	cfg.Options(opts...)
	cfg.Default()

	errCh := make(chan error)

	s.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		res, err := s.cfg.Processor.ProcessHREF(s.collector, href)
		if err != nil {
			if !(errors.Is(err, colly.ErrAlreadyVisited) || errors.Is(err, colly.ErrNoURLFiltersMatch)) {
				errCh <- fmt.Errorf("processing link: %w", err)
			}

			return
		}

		if res.Product == "" {
			return
		}

		if err := cfg.Recorder.RecordProduct(recorder.Product{
			Source: s.cfg.SourceName,
			Name:   res.Product,
			URL:    res.HREF,
		}); err != nil {
			errCh <- fmt.Errorf("recording product: %w", err)
		}
	})

	if err := s.collector.Visit(s.cfg.BaseURL); err != nil {
		return fmt.Errorf("visiting base URL: %w", err)
	}

	go func() {
		s.collector.Wait()

		close(errCh)
	}()

	var finalErr error
	for err := range errCh {
		multierr.AppendInto(&finalErr, err)
	}

	return finalErr
}

type SimpleScraperConfig struct {
	BaseURL    string
	Filters    []*regexp.Regexp
	SourceName string
	Processor  HREFProcessor
}

func (c *SimpleScraperConfig) Options(opts ...SimpleScraperOption) {
	for _, opt := range opts {
		opt.ConfigureSimpleScraper(c)
	}
}

type SimpleScraperOption interface {
	ConfigureSimpleScraper(*SimpleScraperConfig)
}

type HREFProcessor interface {
	ProcessHREF(*colly.Collector, string) (ProcessResult, error)
}

type ProcessResult struct {
	HREF    string
	Product string
}
