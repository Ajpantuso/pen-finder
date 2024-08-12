// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package scraper

import (
	"regexp"

	"github.com/ajpantuso/pen-finder/internal/recorder"
)

type WithBaseURL string

func (w WithBaseURL) ConfigureSimpleScraper(c *SimpleScraperConfig) {
	c.BaseURL = string(w)
}

func (w WithBaseURL) ConfigureSimpleProcessor(c *SimpleProcessorConfig) {
	c.BaseURL = string(w)
}

type WithFilters []*regexp.Regexp

func (w WithFilters) ConfigureSimpleScraper(c *SimpleScraperConfig) {
	c.Filters = append(c.Filters, w...)
}

type WithProductPathPrefix string

func (w WithProductPathPrefix) ConfigureSimpleProcessor(c *SimpleProcessorConfig) {
	c.ProductPathPrefix = string(w)
}

type WithProcessor struct {
	Processor HREFProcessor
}

func (w WithProcessor) ConfigureSimpleScraper(c *SimpleScraperConfig) {
	c.Processor = w.Processor
}

type WithScrapers []Scraper

func (w WithScrapers) ConfigureRun(c *RunConfig) {
	c.Scrapers = append(c.Scrapers, w...)
}

type WithScrapeOptions []ScrapeOption

func (w WithScrapeOptions) ConfigureRun(c *RunConfig) {
	c.ScrapeOptions = append(c.ScrapeOptions, w...)
}

type WithRecorder struct {
	Recorder recorder.Recorder
}

func (w WithRecorder) ConfigureScrape(c *ScrapeConfig) {
	c.Recorder = w.Recorder
}

type WithSourceName string

func (w WithSourceName) ConfigureSimpleScraper(c *SimpleScraperConfig) {
	c.SourceName = string(w)
}
