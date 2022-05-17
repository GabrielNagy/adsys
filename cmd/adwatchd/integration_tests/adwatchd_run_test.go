package adwatchd_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ubuntu/adsys/cmd/adwatchd/commands"
)

func TestRunFailsWhenServiceIsRunning(t *testing.T) {
	t.Parallel()

	var err error
	watchDir := t.TempDir()
	configPath := generateConfig(t, watchDir)

	app := commands.New(commands.WithServiceName("adwatchd-test-1"))
	t.Cleanup(func() {
		uninstallService(t, configPath, &app)
	})

	installService(t, configPath, &app)

	changeAppArgs(t, &app, configPath, "run")
	err = app.Run()
	require.ErrorContains(t, err, "another instance of adwatchd is already running", "Running second instance should fail")
}

func TestRunFailsWhenAnotherInstanceIsRunning(t *testing.T) {
	t.Skip()
	var err error
	watchDir := t.TempDir()
	configPath := generateConfig(t, watchDir)

	app := commands.New(commands.WithServiceName("adwatchd-test-1"))
	changeAppArgs(t, &app, configPath, "run")
	go func() {
		err = app.Run()
		require.NoError(t, err, "Running first instance should succeed")
	}()

	err = app.Run()
	require.ErrorContains(t, err, "another instance of adwatchd is already running", "Running second instance should fail")
}

func TestAppCanQuit(t *testing.T) {
	watchDir := t.TempDir()
	watchDir2 := t.TempDir()
	app := commands.New()
	changeAppArgs(t, &app, "", "run", "--dirs", watchDir, "--dirs", watchDir2)

	wg := sync.WaitGroup{}
	wg.Add(1)
	var err error
	go func() {
		defer wg.Done()
		err = app.Run()
	}()
	app.Quit()

	wg.Wait()
	require.NoError(t, err)
}
