package model

import "time"

// ReloadedSnapshot represents a specific successfully loaded resource revision.
type ReloadedSnapshot[T any] struct {
	Value      T
	LastUpdate time.Time
	Revision   uint64
}

// Reloader represents a reloading resource and provides a snapshot of its current state and the time of the last
// update.
type Reloader[T any] interface {
	// Current returns the current resource snapshot together with its metadata.
	Current() ReloadedSnapshot[T]
	// Snapshot returns a snapshot of the current state of the resource.
	Snapshot() T
	// LastUpdate returns the time of the last successful update to the resource.
	LastUpdate() time.Time
}
