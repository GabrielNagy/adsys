package commands

import (
	"os/exec"
	"strings"
)

func pidCount(name string) (int, error) {
	pids, err := exec.Command("pidof", name).Output()
	if err != nil {
		return 0, err
	}
	return len(strings.Fields(string(pids))), nil
}
