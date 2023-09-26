// Package scripts includes script files that are copied to target VMs and used
// by the e2e test suite.
package scripts

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// Dir returns the directory of the current file.
func Dir() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}
	return filepath.Dir(currentFile), nil
}
