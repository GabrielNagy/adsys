// Package az provides functions to interact with the Azure CLI.
package az

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/adsys/e2e/internal/inventory"
)

// RunCommand runs the Azure CLI with the given arguments.
func RunCommand(ctx context.Context, args ...string) ([]byte, []byte, error) {
	log.Debugf("Running az with args %s", args)

	c := exec.CommandContext(ctx, "az", args...)
	var outb, errb bytes.Buffer
	c.Stdout = &outb
	c.Stderr = &errb
	err := c.Run()

	if outb.Len() > 0 {
		log.Debugf("\tSTDOUT: %s", outb.String())
	}
	if errb.Len() > 0 {
		log.Debugf("\tSTDERR: %s", errb.String())
	}
	return outb.Bytes(), errb.Bytes(), err
}

// DeleteVM deletes the inventory VM in Azure.
func DeleteVM(ctx context.Context, inv inventory.Inventory) error {
	vmName := fmt.Sprintf("adsys-integration-%s-%s", inv.Codename, inv.UUID)
	log.Infof("Deleting VM %q", vmName)

	_, _, err := RunCommand(ctx, "vm", "delete",
		"--resource-group", inv.ResourceGroup,
		"--name", vmName,
		"--force-deletion", "true",
		"--yes",
	)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	return nil
}
