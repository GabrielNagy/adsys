package watcher_test

import (
	"testing"
	"time"
)

func waitForWrites(t *testing.T) {
	t.Helper()

	time.Sleep(time.Millisecond * 100)
}
