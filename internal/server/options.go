// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"time"

	"github.com/ajpantuso/pen-finder/internal/recorder"
	"github.com/go-logr/logr"
)

type WithBindAddr string

func (w WithBindAddr) ConfigureDefaultServer(c *DefaultServerConfig) {
	c.BindAddr = string(w)
}

type WithKeyFile string

func (w WithKeyFile) ConfigureDefaultServer(c *DefaultServerConfig) {
	c.KeyFile = string(w)
}

type WithCertFile string

func (w WithCertFile) ConfigureDefaultServer(c *DefaultServerConfig) {
	c.CertFile = string(w)
}

type WithRunTimeout time.Duration

func (w WithRunTimeout) ConfigureDefaultServer(c *DefaultServerConfig) {
	timeout := time.Duration(w)

	c.RunTimeout = &timeout
}

type WithLogger struct {
	Logger logr.Logger
}

func (w WithLogger) ConfigureDefaultServer(c *DefaultServerConfig) {
	c.Logger = w.Logger
}

type WithRecorder struct {
	Recorder recorder.Recorder
}

func (w WithRecorder) ConfigureDefaultServer(c *DefaultServerConfig) {
	c.Recorder = w.Recorder
}
