package adwatchd_test

import (
	"testing"
	"time"

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
		uninstallService(t, configPath, app)
	})

	installService(t, configPath, app)

	changeAppArgs(t, app, configPath, "run")
	err = app.Run()
	require.ErrorContains(t, err, "another instance of adwatchd is already running", "Running second instance should fail")
}

func TestRunFailsWhenAnotherInstanceIsRunning(t *testing.T) {
	t.Skip()
	var err error
	watchDir := t.TempDir()
	configPath := generateConfig(t, watchDir)

	app := commands.New(commands.WithServiceName("adwatchd-test-1"))
	changeAppArgs(t, app, configPath, "run")
	go func() {
		err = app.Run()
		require.NoError(t, err, "Running first instance should succeed")
	}()

	err = app.Run()
	require.ErrorContains(t, err, "another instance of adwatchd is already running", "Running second instance should fail")
}

// TODO: TestAppCanQuitWithCtrlC.
func TestAppCanQuit(t *testing.T) {
	watchDir := t.TempDir()
	watchDir2 := t.TempDir()
	app := commands.New()
	changeAppArgs(t, app, "", "run", "--dirs", watchDir, "--dirs", watchDir2)

	done := make(chan struct{})
	var err error
	go func() {
		defer close(done)
		err = app.Run()
	}()

	err = app.Quit()
	require.NoError(t, err, "Quitting should succeed")

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("run hasn't exited quickly enough")
	}
	require.NoError(t, err)
}
