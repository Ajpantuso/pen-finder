// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"testing"

	"github.com/ajpantuso/pen-finder/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScraperUnmarshalJSON(t *testing.T) {
	require.Implements(t, new(json.Unmarshaler), new(api.Scraper))

	var scraper api.Scraper

	require.Nil(t, json.Unmarshal([]byte(`"fountain pen hospital"`), &scraper))

	assert.Equal(t, api.ScraperFPH, scraper)
}
