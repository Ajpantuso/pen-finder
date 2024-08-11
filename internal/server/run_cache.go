// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type RunCache interface {
	Get(uuid.UUID) (RunCacheEntry, bool)
	Upsert(uuid.UUID, RunStatus)
}

type RunCacheEntry struct {
	Status      RunStatus
	LastUpdated time.Time
}

type RunStatus string

const (
	RunStatusInProgress RunStatus = "in progress"
	RunStatusSuccess    RunStatus = "success"
	RunStatusFailed     RunStatus = "failed"
)

func NewThreadSafeRunCache() *ThreadSafeRunCache {
	return &ThreadSafeRunCache{
		data: make(map[uuid.UUID]RunCacheEntry),
		lock: &sync.RWMutex{},
	}
}

type ThreadSafeRunCache struct {
	data map[uuid.UUID]RunCacheEntry
	lock *sync.RWMutex
}

func (c *ThreadSafeRunCache) Get(id uuid.UUID) (RunCacheEntry, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	entry, ok := c.data[id]

	return entry, ok
}

func (c *ThreadSafeRunCache) Upsert(id uuid.UUID, status RunStatus) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if entry, ok := c.data[id]; ok && entry.Status == status {
		return
	}

	c.data[id] = RunCacheEntry{
		Status:      status,
		LastUpdated: time.Now(),
	}
}
