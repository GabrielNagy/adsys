package watchdservice_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ubuntu/adsys/cmd/adwatchd/internal/watchdservice"
)

func TestServiceInstallUninstall(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	w, err := watchdservice.New(
		context.Background(),
		watchdservice.AsUserService(),
		watchdservice.WithName("adwatchd-testinstall"),
		watchdservice.WithDirs([]string{temp}))

	require.NoError(t, err, "Could not create service object")

	defer w.Uninstall(context.Background())
	err = w.Install(context.Background())
	require.NoError(t, err, "Could not install service")

	// err = w.Start(context.Background())
	// require.NoError(t, err, "Could not start service")

	// err = w.Restart(context.Background())
	// require.NoError(t, err, "Could not restart service")

	// err = w.Stop(context.Background())
	// require.NoError(t, err, "Could not stop service")

	err = w.Uninstall(context.Background())
	require.NoError(t, err, "Could not uninstall service")
}
