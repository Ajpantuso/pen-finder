// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/ajpantuso/pen-finder/api"
	"github.com/ajpantuso/pen-finder/internal/recorder"
	"github.com/ajpantuso/pen-finder/internal/scraper"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

type Server interface {
	Serve(context.Context) error
}

func NewDefaultServer(opts ...DefaultServerOption) *DefaultServer {
	var cfg DefaultServerConfig

	cfg.Options(opts...)
	cfg.Default()

	return &DefaultServer{
		cfg: cfg,
	}
}

type DefaultServer struct {
	cfg DefaultServerConfig
}

func (s *DefaultServer) Serve(ctx context.Context) error {
	handler := http.NewServeMux()
	handler.HandleFunc("GET /run/{id}", s.handleGetRun)
	handler.HandleFunc("POST /run/", s.handleRunRequest)

	srv := &http.Server{
		Addr:    s.cfg.BindAddr,
		Handler: handler,
	}

	errCh := make(chan error)

	go func() {
		errCh <- srv.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile)
	}()

	for {
		select {
		case err := <-errCh:
			return fmt.Errorf("serving: %w", err)
		case <-ctx.Done():
			s.cfg.Logger.Info("shutting down server")

			return srv.Shutdown(context.Background())
		}
	}
}

func (s *DefaultServer) handleGetRun(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	rawID := r.PathValue("id")

	if rawID == "" {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	runID, err := uuid.Parse(rawID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	entry, found := s.cfg.Cache.Get(runID)
	if !found {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	res := api.GetRunResponse{
		ID:          runID,
		Status:      entry.Status,
		LastUpdated: entry.LastUpdated,
	}

	data, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if _, err := w.Write(data); err != nil {
		s.cfg.Logger.Error(err, "writing response")
	}
}

func (s *DefaultServer) handleRunRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req api.PostRunRequest

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	scrapers := req.Scrapers

	runID := uuid.New()
	res := api.PostRunResponse{
		RunID: runID,
	}

	data, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	_, err = w.Write(data)
	if err != nil {
		s.cfg.Logger.Error(err, "writing response")
	}

	s.cfg.Cache.Upsert(runID, api.RunStatusInProgress)

	ctx := context.Background()
	cancel := func() {}

	if s.cfg.RunTimeout != nil {
		ctx, cancel = context.WithTimeout(ctx, *s.cfg.RunTimeout)
	}
	defer cancel()

	var scrapeOpts []scraper.ScrapeOption
	if s.cfg.Recorder != nil {
		scrapeOpts = append(scrapeOpts, scraper.WithRecorder{Recorder: s.cfg.Recorder})
	}

	s.cfg.Logger.Info("running scrappers", "runID", runID, "scrapers", scrapers)
	go func() {
		if err := s.cfg.Runner.Run(ctx, scraper.WithScrapers(mapScrapers(scrapers)), scraper.WithScrapeOptions(scrapeOpts)); err != nil {
			s.cfg.Cache.Upsert(runID, api.RunStatusFailed)
			s.cfg.Logger.Error(err, "running scrapers")
		}

		s.cfg.Cache.Upsert(runID, api.RunStatusSuccess)
	}()
}

func mapScrapers(scrapers []api.Scraper) []scraper.Scraper {
	if len(scrapers) < 1 {
		return []scraper.Scraper{
			newChatterlyScraper(),
			newFPHScraper(),
			newTruphaeScraper(),
		}
	}

	slices.Sort(scrapers)
	scrapers = slices.Compact(scrapers)

	result := make([]scraper.Scraper, 0, len(scrapers))
	for _, s := range scrapers {
		switch s {
		case api.ScraperChatterly:
			result = append(result, newChatterlyScraper())
		case api.ScraperFPH:
			result = append(result, newFPHScraper())
		case api.ScraperTruphae:
			result = append(result, newTruphaeScraper())
		default:
			continue
		}
	}

	return result
}

type DefaultServerConfig struct {
	Runner     scraper.Runner
	BindAddr   string
	KeyFile    string
	CertFile   string
	RunTimeout *time.Duration
	Logger     logr.Logger
	Cache      RunCache
	Recorder   recorder.Recorder
}

func (c *DefaultServerConfig) Options(opts ...DefaultServerOption) {
	for _, opt := range opts {
		opt.ConfigureDefaultServer(c)
	}
}

func (c *DefaultServerConfig) Default() {
	if c.Logger.GetSink() == nil {
		c.Logger = logr.Discard()
	}
	if c.Cache == nil {
		c.Cache = NewThreadSafeRunCache()
	}
	if c.Runner == nil {
		c.Runner = scraper.NewParallelRunner()
	}
}

type DefaultServerOption interface {
	ConfigureDefaultServer(*DefaultServerConfig)
}

func newChatterlyScraper() *scraper.SimpleScraper {
	return scraper.NewSimpleScraper(
		scraper.WithBaseURL("https://chatterleyluxuries.com/product-category/pens/consignments"),
		scraper.WithFilters{
			regexp.MustCompile(`https://chatterleyluxuries\.com/product-category/pens/consignments.*`),
			regexp.MustCompile(`https://chatterleyluxuries\.com/product/.*`),
		},
		scraper.WithSourceName("chatterly_luxuries"),
		scraper.WithProcessor{Processor: scraper.NewSimpleProcessor(
			scraper.WithBaseURL("https://chatterleyluxuries.com"),
			scraper.WithProductPathPrefix("/product/"),
		)})
}

func newFPHScraper() *scraper.SimpleScraper {
	return scraper.NewSimpleScraper(
		scraper.WithBaseURL("https://fountainpenhospital.com/collections/back-room-1"),
		scraper.WithFilters{regexp.MustCompile(`https://fountainpenhospital\.com/collections/back-room-1.*`)},
		scraper.WithSourceName("chatterly_luxuries"),
		scraper.WithProcessor{Processor: scraper.NewSimpleProcessor(
			scraper.WithBaseURL("https://fountainpenhospital.com"),
			scraper.WithProductPathPrefix("/collections/back-room-1/products/"),
		)})
}

func newTruphaeScraper() *scraper.SimpleScraper {
	return scraper.NewSimpleScraper(
		scraper.WithBaseURL("https://truphaeinc.com/collections/pre-owned-pens"),
		scraper.WithFilters{regexp.MustCompile(`https://truphaeinc\.com/collections/pre-owned-pens.*`)},
		scraper.WithSourceName("truphae"),
		scraper.WithProcessor{Processor: scraper.NewSimpleProcessor(
			scraper.WithBaseURL("https://truphaeinc.com/"),
			scraper.WithProductPathPrefix("/collections/pre-owned-pens/products/"),
		)})
}
