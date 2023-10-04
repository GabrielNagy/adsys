// Package main provides a script to generalize an Azure VM to be used as a
// template for integration tests.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/adsys/e2e/internal/command"
	"github.com/ubuntu/adsys/e2e/internal/inventory"
	"github.com/ubuntu/adsys/e2e/scripts"
)

var codename string
var preserve bool

func main() {
	os.Exit(run())
}

func run() int {
	cmd := command.New(action,
		command.WithValidateFunc(validate),
		command.WithStateTransition(inventory.Null, inventory.PackageBuilt),
	)
	cmd.Usage = fmt.Sprintf(`go run ./%s [options]

Generalize an Azure VM to use as a template for integration tests.

Options:
 --codename       Required: codename of the Ubuntu release to build for (e.g. focal)
 -p, --preserve   Don't delete the build container after finishing (default: false)

This script will:
 - build the adsys package in a Docker container from the current source tree for the given codename
`, filepath.Base(os.Args[0]))

	cmd.AddStringFlag(&codename, "codename", "", "")
	cmd.AddBoolFlag(&preserve, "preserve", false, "")
	cmd.AddBoolFlag(&preserve, "p", false, "")

	return cmd.Execute(context.Background())
}

func validate(_ context.Context, _ *command.Command) error {
	if codename == "" {
		return errors.New("codename is required")

	}
	return nil
}

func action(ctx context.Context, cmd *command.Command) error {
	dockerTag := fmt.Sprintf("adsys-build-%s:latest", codename)

	scriptsDir, err := scripts.Dir()
	if err != nil {
		return err
	}

	log.Infof("Preparing build container %q", dockerTag)
	// Build the container
	out, err := exec.Command(
		"docker", "build", "-t", dockerTag,
		"--build-arg", fmt.Sprintf("CODENAME=%s", codename),
		"--file", filepath.Join(scriptsDir, "Dockerfile.build"), ".",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build container: %w: %s", err, string(out))
	}
	log.Debugf("docker build output: %s", string(out))

	// Run the container
	dockerArgs := []string{"run"}
	if !preserve {
		dockerArgs = append(dockerArgs, "--rm")
	}

	adsysRootDir, err := scripts.RootDir()
	if err != nil {
		return err
	}
	dockerArgs = append(dockerArgs,
		"-v", fmt.Sprintf("%s:/source-ro:ro", adsysRootDir),
		"-v", fmt.Sprintf("%s/output:/output", adsysRootDir),
		"-v", fmt.Sprintf("%s/build-deb.sh:/build-helper.sh:ro", scriptsDir),
		"-v", fmt.Sprintf("%s/patches:/patches:ro", scriptsDir),
		// This is to set correct permissions on the output directory
		"-e", fmt.Sprintf("USER=%d", os.Getuid()),
		"-e", fmt.Sprintf("GROUP=%d", os.Getgid()),
		"--tmpfs", "/tmp:exec",
		dockerTag,
		"/build-helper.sh",
	)

	log.Info("Building adsys package")
	log.Debugf("Running docker with args: %v", dockerArgs)
	out, err = exec.Command("docker", dockerArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run container: %w: %s", err, string(out))
	}

	cmd.Inventory.Codename = codename

	return nil
}
