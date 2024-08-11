// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScraperUnmarshalJSON(t *testing.T) {
	require.Implements(t, new(json.Unmarshaler), new(Scraper))

	var scraper Scraper

	require.Nil(t, json.Unmarshal([]byte(`"fountain pen hospital"`), &scraper))

	assert.Equal(t, ScraperFPH, scraper)
}
