// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type GetRunResponse struct {
	ID          uuid.UUID `json:"id"`
	Status      RunStatus `json:"status"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type RunStatus string

const (
	RunStatusInProgress RunStatus = "in progress"
	RunStatusSuccess    RunStatus = "success"
	RunStatusFailed     RunStatus = "failed"
)

type PostRunRequest struct {
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
	ScraperTruphae   Scraper = "truphae"
)

type PostRunResponse struct {
	RunID uuid.UUID `json:"runID"`
}
