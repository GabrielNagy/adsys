package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func waitForWrites(t *testing.T, dirs ...string) {
	t.Helper()

	// Windows doesn't have a syscall.Sync function, so the next best thing to
	// do is to force a walk of the directory to make sure the writes are picked
	// up.
	//
	// Otherwise the watcher could detect changes just as soon as it starts
	// walking paths.
	for _, dir := range dirs {
		filepath.WalkDir(dir, func(_ string, _ os.DirEntry, _ error) error { return nil })
	}
	time.Sleep(time.Millisecond * 100)
}
