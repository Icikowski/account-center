package catalog

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xcopy"
	"git.sr.ht/~icikowski/account-center/internal/shared/xlog"
)

type catalogWatcher struct {
	mux        sync.RWMutex
	snapshot   *model.Catalog
	lastUpdate time.Time
	revision   uint64
	path       string
	log        zerolog.Logger
}

// NewWatcher creates a new [model.Catalog] watcher that monitors the specified
// file for changes and reloads the data accordingly.
func NewWatcher(
	ctx context.Context,
	path string,
	debounce time.Duration,
	log zerolog.Logger,
) (model.Reloader[model.Catalog], error) {
	w := &catalogWatcher{
		path: path,
		log:  log,
	}

	if err := w.fetch(); err != nil {
		return nil, err
	}

	if err := w.watch(ctx, debounce); err != nil {
		return nil, err
	}
	return w, nil
}

// Snapshot implements [model.Reloader].
func (w *catalogWatcher) Snapshot() model.Catalog {
	return w.Current().Value
}

// LastUpdate implements [model.Reloader].
func (w *catalogWatcher) LastUpdate() time.Time {
	return w.Current().LastUpdate
}

// Current implements [model.Reloader].
func (w *catalogWatcher) Current() model.ReloadedSnapshot[model.Catalog] {
	w.mux.RLock()
	defer w.mux.RUnlock()

	catalogCopy, err := xcopy.DeepCopy(w.snapshot)
	if err != nil {
		w.log.Error().Err(err).Msg("failed to deep copy catalog snapshot")
		return model.ReloadedSnapshot[model.Catalog]{}
	}

	return model.ReloadedSnapshot[model.Catalog]{
		Value:      *catalogCopy,
		LastUpdate: w.lastUpdate,
		Revision:   w.revision,
	}
}

func (w *catalogWatcher) fetch() error {
	catalog, err := Load(w.path)
	if err != nil {
		return err
	}

	w.mux.Lock()
	defer w.mux.Unlock()

	w.snapshot = catalog
	w.lastUpdate = time.Now()
	w.revision++
	return nil
}

func (w *catalogWatcher) watch(ctx context.Context, debounce time.Duration) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	targetPath := filepath.Clean(w.path)
	watchPath := filepath.Dir(targetPath)

	debounce = max(debounce, 100*time.Millisecond)

	if err := watcher.Add(watchPath); err != nil {
		_ = watcher.Close()
		return err
	}

	go func() {
		defer func() {
			if err := watcher.Close(); err != nil {
				w.log.Error().Err(err).Msg("failed to close catalog watcher")
			}
		}()

		var (
			timer   *time.Timer
			timerCh <-chan time.Time
		)

		for {
			select {
			case <-ctx.Done():
				return
			case <-timerCh:
				timerCh = nil
				if timer != nil {
					timer.Stop()
				}

				if err := w.fetch(); err != nil {
					w.log.Error().Err(err).Str(xlog.FieldPath, w.path).Msg("failed to reload catalog")
				} else {
					w.log.Info().Str(xlog.FieldPath, w.path).Msg("catalog reloaded")
				}
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if filepath.Clean(event.Name) != targetPath {
					continue
				}

				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
					continue
				}

				if timer == nil {
					timer = time.NewTimer(debounce)
				} else {
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(debounce)
				}
				timerCh = timer.C
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				w.log.Error().Err(err).Str(xlog.FieldPath, w.path).Msg("catalog watcher error")
			}
		}
	}()

	return nil
}
