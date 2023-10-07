// Package main provides a script to generalize an Azure VM to be used as a
// template for integration tests.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/maruel/natural"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/adsys/e2e/internal/az"
	"github.com/ubuntu/adsys/e2e/internal/command"
	"github.com/ubuntu/adsys/e2e/internal/inventory"
)

var version string
var preserve bool

func main() {
	os.Exit(run())
}

func run() int {
	cmd := command.New(action, command.WithStateTransition(inventory.BaseVMCreated, inventory.TemplateCreated))
	cmd.Usage = fmt.Sprintf(`go run ./%s [options]

Generalize an Azure VM to use as a template for integration tests.

Options:
 --version          override the template version number (default behavior is to
                    auto-increment the latest version by 0.0.1)
 -p, --preserve     preserve base VM after creating image version (default: false)

This script will:
 - create an Azure image definition for the Ubuntu version of the VM unless it already exists
 - create an image version using the VM, incrementing the version number
 - destroy the base VM unless otherwise specified

The script requires an inventory file to be present in the current directory,
created by the 00_prepare_base_vm script.

The machine must be authenticated to Azure via the Azure CLI.`, filepath.Base(os.Args[0]))

	cmd.AddStringFlag(&version, "version", "", "")
	cmd.AddBoolFlag(&preserve, "preserve", false, "")

	return cmd.Execute(context.Background())
}

func action(ctx context.Context, cmd *command.Command) error {
	inv := cmd.Inventory

	imageDefinition := az.ImageDefinitionName(inv.Codename)
	latestImageVersion, err := az.LatestImageVersion(ctx, imageDefinition)
	if err != nil {
		return err
	}

	isDevelopmentVersion := strings.Contains(cmd.Inventory.BaseVMImage, "daily")
	nextImageVersion := incrementVersion(latestImageVersion, isDevelopmentVersion)

	// Destroy VM if template creation fails
	defer func() {
		if err == nil {
			return
		}
		log.Error(err)

		if preserve {
			log.Infof("Preserving VM as requested...")
			return
		}

		if err := az.DeleteVM(context.Background(), cmd.Inventory.VMName); err != nil {
			log.Error(err)
		}
	}()

	// If the version is empty, we need to create the image definition
	if latestImageVersion == "" {
		log.Infof("Creating image definition %q", imageDefinition)
		_, _, err := az.RunCommand(ctx, "sig", "image-definition", "create",
			"--resource-group", "AD",
			"--gallery-name", "AD",
			"--gallery-image-definition", imageDefinition,
			"--publisher", "Canonical",
			"--offer", imageDefinition,
			"--sku", inv.Codename,
			"--os-type", "Linux",
			"--os-state", "Specialized",
			"--hyper-v-generation", "V2",
			"--features", "SecurityType=TrustedLaunch",
			"--tags", "project=AD", "subproject=adsys-integration-tests",
		)
		if err != nil {
			return fmt.Errorf("failed to create image definition: %w", err)
		}
	}

	// User has specified a version, use it instead
	if version != "" {
		nextImageVersion = version
	}

	// Create the image version
	log.Infof("Creating image version %q for image definition %q", nextImageVersion, imageDefinition)
	_, _, err = az.RunCommand(ctx, "sig", "image-version", "create",
		"--resource-group", "AD",
		"--gallery-name", "AD",
		"--gallery-image-definition", imageDefinition,
		"--gallery-image-version", nextImageVersion,
		"--target-regions", "westeurope", "eastus=1=standard_zrs",
		"--replica-count", "2",
		"--managed-image", inv.VMID,
		"--tags", "project=AD", "subproject=adsys-integration-tests",
	)
	if err != nil {
		return fmt.Errorf("failed to create image version: %w", err)
	}

	// Destroy base VM unless otherwise specified
	if preserve {
		log.Infof("Preserving resource %q as requested", inv.VMID)
		return nil
	}
	if err := az.DeleteVM(ctx, cmd.Inventory.VMName); err != nil {
		return err
	}

	return nil
}

func incrementVersion(version string, dev bool) string {
	firstVersion := "0.0.1"
	// Non-development images begin at 1.0.0
	if !dev {
		firstVersion = "1.0.0"
		if natural.Less(version, firstVersion) {
			return firstVersion
		}
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return firstVersion
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return firstVersion
	}
	patch++

	return fmt.Sprintf("%s.%d", strings.Join(parts[:2], "."), patch)
}
