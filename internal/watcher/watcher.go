package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kardianos/service"
	"github.com/ubuntu/adsys/internal/decorate"
	log "github.com/ubuntu/adsys/internal/grpc/logstreamer"
	"github.com/ubuntu/adsys/internal/i18n"
	"gopkg.in/ini.v1"
)

const (
	gptFileName = "gpt.ini"
)

// Watcher provides options necessary to watch a directory and its children.
type Watcher struct {
	dirs      []string
	parentCtx context.Context
	cancel    context.CancelFunc
	watching  chan struct{}

	refreshDuration time.Duration
}

type options struct {
	refreshDuration time.Duration
}
type option func(*options) error

// New returns a new Watcher instance.
func New(ctx context.Context, dirs []string, opts ...option) (*Watcher, error) {
	if len(dirs) == 0 {
		return nil, errors.New(i18n.G("no directories to watch"))
	}

	args := options{
		refreshDuration: 10 * time.Second,
	}
	// applied options
	for _, o := range opts {
		if err := o(&args); err != nil {
			return nil, err
		}
	}

	return &Watcher{
		dirs: dirs,

		parentCtx:       ctx,
		refreshDuration: args.refreshDuration,
	}, nil
}

// Dirs returns the directories currently being watched.
func (w *Watcher) Dirs() []string {
	return w.dirs
}

// Start is called by the service manager to start the watcher service.
func (w *Watcher) Start(s service.Service) (err error) {
	decorate.OnError(&err, i18n.G("can't start service"))

	return w.startWatch(w.parentCtx, w.dirs)
}

func (w *Watcher) startWatch(ctx context.Context, dirs []string) error {
	ctx, cancel := context.WithCancel(w.parentCtx)
	w.cancel = cancel

	w.watching = make(chan struct{})
	initError := make(chan error)
	go func() {
		if errWatching := w.watch(ctx, w.dirs, initError); errWatching != nil {
			log.Warningf(ctx, "Watch failed: %v", errWatching)
		}
	}()
	return <-initError
}

// Stop is called by the service manager to stop the watcher service.
func (w *Watcher) Stop(s service.Service) (err error) {
	decorate.OnError(&err, i18n.G("can't stop service"))

	return w.stopWatch(w.parentCtx)
}

func (w *Watcher) stopWatch(ctx context.Context) error {
	if w.cancel == nil {
		return errors.New(i18n.G("the service is already stopping or not running"))
	}

	w.cancel()
	w.cancel = nil

	// wait for watching to be closed
	for {
		_, ok := <-w.watching
		if ok {
			continue
		}
		break
	}

	return nil
}

// UpdateDirs restarts watch loop with new directories.
func (w *Watcher) UpdateDirs(dirs []string) (err error) {
	decorate.OnError(&err, i18n.G("can't update directories to watch"))
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf(i18n.G("directory %v does not exist"), dir)
		}
	}

	log.Debugf(w.parentCtx, "Updating directories to %v", dirs)

	if err := w.stopWatch(w.parentCtx); err != nil {
		return err
	}

	w.dirs = dirs
	return w.startWatch(w.parentCtx, dirs)
}

func (w *Watcher) watch(ctx context.Context, dirs []string, initError chan<- error) (err error) {
	decorate.OnError(&err, i18n.G("can't watch over %v"), dirs)
	defer close(w.watching)

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		initError <- err
	}
	defer fsWatcher.Close()

	// Collect directories to watch.
	for _, dir := range dirs {
		if err := watchSubDirs(ctx, fsWatcher, dir); err != nil {
			initError <- err
		}
	}

	// We have a grace period of 10s without any changes before committing the changes.
	refreshTimer := time.NewTimer(w.refreshDuration)
	defer refreshTimer.Stop()
	refreshTimer.Stop()

	// Collect directories to watch.
	var modifiedRootDirs []string
	initError <- nil
	for {
		select {
		case event, ok := <-fsWatcher.Events:
			if !ok {
				continue
			}
			log.Debugf(ctx, "Got event: %v", event)

			// If the modified file is our changes, ignore.
			if strings.ToLower(filepath.Base(event.Name)) == gptFileName {
				continue
			}

			// Add new detected files and directories for content to watch.
			if event.Op&fsnotify.Create == fsnotify.Create {
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					log.Warningf(ctx, "Failed to stat: %s", err)
					continue
				}

				if fileInfo.IsDir() {
					if err := watchSubDirs(ctx, fsWatcher, event.Name); err != nil {
						return err
					}
				} else if fileInfo.Mode().IsRegular() {
					fsWatcher.Add(event.Name)
				}
			}

			// Remove deleted or renamed files/directories from the watch list.
			if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				fsWatcher.Remove(event.Name)
			}

			// Check there is something to update
			if !(event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create || // Rename is always followed by a Create.
				event.Op&fsnotify.Remove == fsnotify.Remove) {
				continue
			}

			// Find and add matching root directory if not already present in the list to refresh.
			rootDir, err := getRootDir(event.Name, dirs)
			if err != nil {
				log.Warningf(ctx, "%v", err)
				continue
			}
			var alreadyAdded bool
			for _, modifiedRootDir := range modifiedRootDirs {
				if rootDir != modifiedRootDir {
					continue
				}
				alreadyAdded = true
				break
			}
			if !alreadyAdded {
				modifiedRootDirs = append(modifiedRootDirs, rootDir)
			}

			// Set the grace period of 10s without any changes before bumping the version.
			// Stop means that the timer expired, not that it was stopped, so drain the channel only if there is something to drain.
			if !refreshTimer.Stop() {
				select {
				case <-refreshTimer.C:
				default:
				}
			}
			refreshTimer.Reset(w.refreshDuration)

		case err, ok := <-fsWatcher.Errors:
			if ok {
				log.Warningf(ctx, "Got event error: %v", err)
			}
			continue

		case <-refreshTimer.C:
			// Updating GPT.ini.
			updateVersions(ctx, modifiedRootDirs)

		case <-ctx.Done():
			log.Infof(ctx, "Watcher stopped")
			// Check if there was a timer in progress to not miss an update before exiting.
			if refreshTimer.Stop() {
				updateVersions(ctx, modifiedRootDirs)
			}

			return nil
		}
	}
}

func watchSubDirs(ctx context.Context, fsWatcher *fsnotify.Watcher, path string) (err error) {
	decorate.OnError(&err, i18n.G("can't watch directory and children of %s"), path)
	log.Debugf(ctx, "Watching %s and children", path)

	err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		log.Debugf(ctx, "Watching: %v", p)
		return fsWatcher.Add(p)
	})
	return err
}

func getRootDir(path string, rootDirs []string) (string, error) {
	var rootDir string
	var currentRootDirLength int
	for _, root := range rootDirs {
		if strings.HasPrefix(path, root) {
			if len(root) <= currentRootDirLength {
				continue
			}
			rootDir = root
			currentRootDirLength = len(root)
		}
	}
	if rootDir == "" {
		return "", fmt.Errorf(i18n.G("no root directory matching %s found"), path)
	}

	return rootDir, nil
}

func updateVersions(ctx context.Context, modifiedRootDirs []string) {
	for _, dir := range modifiedRootDirs {
		gptIniPath := filepath.Join(dir, gptFileName)
		if err := bumpVersion(ctx, gptIniPath); err != nil {
			log.Warningf(ctx, "Failed to bump %s version: %s", gptIniPath, err)
		}
	}
}

func bumpVersion(ctx context.Context, path string) (err error) {
	decorate.OnError(&err, i18n.G("can't bump version for %s"), path)
	log.Infof(ctx, "Bumping version for %s", path)

	cfg, err := ini.Load(path)
	if err != nil {
		log.Warningf(ctx, "error loading ini contents: %v, creating a new file", err)
		cfg = ini.Empty()
		cfg.Section("General").NewKey("Version", "0")
	}

	v, err := cfg.Section("General").Key("Version").Int()
	if err != nil {
		return err
	}

	v++
	cfg.Section("General").Key("Version").SetValue(strconv.Itoa(v))

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = cfg.WriteTo(f)

	return err
}
