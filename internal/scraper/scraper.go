// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package scraper

import (
	"context"
	"github.com/ajpantuso/pen-finder/internal/recorder"
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

type WithRecorder struct {
	Recorder recorder.Recorder
}

func (w WithRecorder) ConfigureScrape(c *ScrapeConfig) {
	c.Recorder = w.Recorder
}

// TODO: implement Bertram's scraper
// TODO: implement Fahrney's scraper
// TODO: implement Dromgoole's scraper
// TODO: implement Truphae scraper
// TODO: implement ChatterlyLuxuries scraper
// TODO: implement PenReam scraper
// TODO: implement ThePenMarket scraper
// TODO: implement ItalianPens scraper
// TODO: implement Etsy scraper
// TODO: implement eBay scraper
