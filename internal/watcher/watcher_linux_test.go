package watcher_test

import (
	"syscall"
	"testing"
	"time"
)

func waitForWrites(t *testing.T) {
	t.Helper()

	// Give time for the writes to be picked up
	syscall.Sync()
	time.Sleep(time.Millisecond * 100)
}
