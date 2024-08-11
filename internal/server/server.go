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

	"github.com/ajpantuso/pen-finder/internal/recorder"
	"github.com/ajpantuso/pen-finder/internal/scraper"
	"github.com/ajpantuso/pen-finder/internal/scraper/chatterley"
	"github.com/ajpantuso/pen-finder/internal/scraper/fph"
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

	res := GetRunResponse{
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

type GetRunResponse struct {
	ID          uuid.UUID `json:"id"`
	Status      RunStatus `json:"status"`
	LastUpdated time.Time `json:"lastUpdated"`
}

func (s *DefaultServer) handleRunRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req RunRequest

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	scrapers := req.Scrapers

	runID := uuid.New()
	res := RunResponse{
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

	s.cfg.Cache.Upsert(runID, RunStatusInProgress)

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

	s.cfg.Logger.Info("running scrappers", "runID", runID)
	go func() {
		if err := s.cfg.Runner.Run(ctx, scraper.WithScrapers(mapScrapers(scrapers)), scraper.WithScrapeOptions(scrapeOpts)); err != nil {
			s.cfg.Cache.Upsert(runID, RunStatusFailed)
			s.cfg.Logger.Error(err, "running scrapers")
		}

		s.cfg.Cache.Upsert(runID, RunStatusSuccess)
	}()
}

func mapScrapers(scrapers []Scraper) []scraper.Scraper {
	if len(scrapers) < 1 {
		return []scraper.Scraper{
			fph.NewScraper(
				"https://fountainpenhospital.com/collections/back-room-1",
				[]*regexp.Regexp{
					regexp.MustCompile(`https://fountainpenhospital\.com/collections/back-room-1.*`),
				},
			),
			chatterley.NewScraper(
				"https://chatterleyluxuries.com/product-category/pens/consignments",
				[]*regexp.Regexp{
					regexp.MustCompile(`https://chatterleyluxuries\.com/product-category/pens/consignments.*`),
					regexp.MustCompile(`https://chatterleyluxuries\.com/product/.*`),
				},
			),
		}
	}

	slices.Sort(scrapers)
	scrapers = slices.Compact(scrapers)

	result := make([]scraper.Scraper, 0, len(scrapers))
	for _, s := range scrapers {
		switch s {
		case ScraperFPH:
			result = append(result, fph.NewScraper(
				"https://fountainpenhospital.com/collections/back-room-1",
				[]*regexp.Regexp{
					regexp.MustCompile(`https://fountainpenhospital\.com/collections/back-room-1.*`),
				},
			),
			)
		case ScraperChatterly:
			result = append(result, chatterley.NewScraper(
				"https://chatterleyluxuries.com/product-category/pens/consignments",
				[]*regexp.Regexp{
					regexp.MustCompile(`https://chatterleyluxuries\.com/category/pens/consignments/.*`),
					regexp.MustCompile(`https://chatterleyluxuries\.com/product/.*`),
				},
			),
			)
		default:
			continue
		}
	}

	return result
}

type RunRequest struct {
	Scrapers []Scraper `json:"scrapers"`
}

type Scraper string

func (s *Scraper) UnmarshalJSON(data []byte) error {
	if n := len(data); n > 1 && (data[0] == '"' && data[n-1] == '"') {
		return json.Unmarshal(data, (*string)(s))
	}
	switch Scraper(data) {
	case ScraperFPH:
		*s = ScraperFPH
	default:
		*s = ScraperUnknown
	}

	return nil
}

const (
	ScraperUnknown   Scraper = "unknown"
	ScraperFPH       Scraper = "fountain pen hospital"
	ScraperChatterly Scraper = "chatterly luxuries"
)

type RunResponse struct {
	RunID uuid.UUID `json:"runID"`
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
