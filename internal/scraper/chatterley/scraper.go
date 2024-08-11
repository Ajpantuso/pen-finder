// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package chatterley

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ajpantuso/pen-finder/internal/recorder"
	"github.com/ajpantuso/pen-finder/internal/scraper"
	"github.com/gocolly/colly/v2"
	"go.uber.org/multierr"
)

func NewScraper(base string, filters []*regexp.Regexp) *Scraper {
	return &Scraper{
		collector: colly.NewCollector(colly.Async(), colly.URLFilters(filters...)),
		base:      base,
	}
}

type Scraper struct {
	collector *colly.Collector
	base      string
}

func (s *Scraper) Scrape(ctx context.Context, opts ...scraper.ScrapeOption) error {
	var cfg scraper.ScrapeConfig

	cfg.Options(opts...)
	cfg.Default()

	errCh := make(chan error)

	s.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		res, err := s.processHREF(href)
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
			Source: "chatterly_luxuries",
			Name:   res.Product,
			URL:    res.HREF,
		}); err != nil {
			errCh <- fmt.Errorf("recording product: %w", err)
		}
	})

	if err := s.collector.Visit(s.base); err != nil {
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

func (s *Scraper) processHREF(href string) (result, error) {
	link := href
	if !strings.HasPrefix(link, "/") {
		if err := s.collector.Visit(link); err != nil {
			return result{}, fmt.Errorf("visiting %s: %w", link, err)
		}
	} else {
		var err error
		link, err = url.JoinPath("https://chatterleyluxuries.com", href)
		if err != nil {
			return result{}, fmt.Errorf("joining path: %w", err)
		}

		if err := s.collector.Visit(link); err != nil {
			return result{}, fmt.Errorf("visiting %s: %w", link, err)
		}
	}

	var product string
	url, err := url.Parse(link)
	if err != nil {
		return result{}, fmt.Errorf("parsing link: %w", err)
	}

	if strings.HasPrefix(url.Path, "/product/") {
		product = strings.TrimSuffix(strings.TrimPrefix(url.Path, "/product/"), "/")
	}

	return result{
		HREF:    link,
		Product: product,
	}, nil
}

type result struct {
	HREF    string
	Product string
}
