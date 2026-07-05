package model

import "time"

// Evaluation represents a cached access-control evaluation result.
type Evaluation struct {
	CatalogTimestamp  time.Time
	Groups            []string
	EffectiveServices []EffectiveService
}
