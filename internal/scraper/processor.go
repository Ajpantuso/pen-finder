// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package scraper

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

func NewSimpleProcessor(opts ...SimpleProcessorOption) *SimpleProcessor {
	var cfg SimpleProcessorConfig

	cfg.Options(opts...)

	return &SimpleProcessor{
		cfg: cfg,
	}
}

type SimpleProcessor struct {
	cfg SimpleProcessorConfig
}

type SimpleProcessorConfig struct {
	BaseURL           string
	ProductPathPrefix string
}

func (c *SimpleProcessorConfig) Options(opts ...SimpleProcessorOption) {
	for _, opt := range opts {
		opt.ConfigureSimpleProcessor(c)
	}
}

type SimpleProcessorOption interface {
	ConfigureSimpleProcessor(*SimpleProcessorConfig)
}

func (p *SimpleProcessor) ProcessHREF(collector *colly.Collector, href string) (ProcessResult, error) {
	link := href
	if !strings.HasPrefix(link, "/") {
		if err := collector.Visit(link); err != nil {
			return ProcessResult{}, fmt.Errorf("visiting %s: %w", link, err)
		}
	} else {
		var err error
		link, err = url.JoinPath(p.cfg.BaseURL, href)
		if err != nil {
			return ProcessResult{}, fmt.Errorf("joining path: %w", err)
		}

		if err := collector.Visit(link); err != nil {
			return ProcessResult{}, fmt.Errorf("visiting %s: %w", link, err)
		}
	}

	var product string
	url, err := url.Parse(link)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("parsing link: %w", err)
	}

	if strings.HasPrefix(url.Path, p.cfg.ProductPathPrefix) {
		product = strings.TrimSuffix(strings.TrimPrefix(url.Path, p.cfg.ProductPathPrefix), "/")
	}

	return ProcessResult{
		HREF:    link,
		Product: product,
	}, nil
}
