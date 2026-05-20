package knowledgebase

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xcopy"
)

type knowledgeBaseWatcher struct {
	mux        sync.RWMutex
	snapshot   *model.KnowledgeBase
	lastUpdate time.Time
	revision   uint64
	path       string
	log        zerolog.Logger
}

// NewWatcher creates a new [model.KnowledgeBase] watcher that monitors the specified
// directory for changes and reloads the data accordingly.
func NewWatcher(
	ctx context.Context,
	path string,
	debounce time.Duration,
	log zerolog.Logger,
) (model.Reloader[model.KnowledgeBase], error) {
	w := &knowledgeBaseWatcher{
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
func (w *knowledgeBaseWatcher) Snapshot() model.KnowledgeBase {
	return w.Current().Value
}

// LastUpdate implements [model.Reloader].
func (w *knowledgeBaseWatcher) LastUpdate() time.Time {
	return w.Current().LastUpdate
}

// Current implements [model.Reloader].
func (w *knowledgeBaseWatcher) Current() model.ReloadedSnapshot[model.KnowledgeBase] {
	w.mux.RLock()
	defer w.mux.RUnlock()

	knowledgeBaseCopy, err := xcopy.DeepCopy(w.snapshot)
	if err != nil {
		w.log.Error().Err(err).Msg("failed to deep copy knowledge base snapshot")
		return model.ReloadedSnapshot[model.KnowledgeBase]{}
	}

	return model.ReloadedSnapshot[model.KnowledgeBase]{
		Value:      *knowledgeBaseCopy,
		LastUpdate: w.lastUpdate,
		Revision:   w.revision,
	}
}

func (w *knowledgeBaseWatcher) fetch() error {
	knowledgeBase, err := Load(w.path)
	if err != nil {
		return err
	}

	w.mux.Lock()
	defer w.mux.Unlock()

	w.snapshot = knowledgeBase
	w.lastUpdate = time.Now()
	w.revision++
	return nil
}

func (w *knowledgeBaseWatcher) watch(ctx context.Context, debounce time.Duration) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := addDirectoryWatches(watcher, filepath.Clean(w.path)); err != nil {
		_ = watcher.Close()
		return err
	}

	debounce = max(debounce, 100*time.Millisecond)

	go func() {
		defer func() {
			if err := watcher.Close(); err != nil {
				w.log.Error().Err(err).Msg("failed to close knowledge base watcher")
			}
		}()

		var (
			timer   *time.Timer
			timerCh <-chan time.Time
		)

		triggerReload := func() {
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
		}

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
					w.log.Error().
						Err(err).
						Str("path", w.path).
						Msg("failed to reload knowledge base")
				} else {
					w.log.Info().Str("path", w.path).Msg("knowledge base reloaded")
				}
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) == 0 {
					continue
				}

				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						if err := addDirectoryWatches(
							watcher,
							filepath.Clean(event.Name),
						); err != nil {
							w.log.Error().
								Err(err).
								Str("path", event.Name).
								Msg("failed to watch knowledge base directory")
						}
					}
				}

				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					_ = watcher.Remove(event.Name)
				}

				triggerReload()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				w.log.Error().Err(err).Str("path", w.path).Msg("knowledge base watcher error")
			}
		}
	}()

	return nil
}

func addDirectoryWatches(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(currentPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		return watcher.Add(currentPath)
	})
}
