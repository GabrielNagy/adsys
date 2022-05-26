package interactive_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
	"github.com/ubuntu/adsys/cmd/adwatchd/interactive"
)

var (
	update bool
	stdout bool
)

func TestInteractiveInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		events        []tea.Msg
		existingPaths []string
	}{
		"write something": {
			events: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("foo/baz")},
				tea.KeyMsg{Type: tea.KeyEnter},
				tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("foo/bar")},
				tea.KeyMsg{Type: tea.KeyDown},
				// tea.KeyMsg{Type: tea.KeyEnter},
				//{Type: tea.KeyEnter},
				//
			},
			existingPaths: []string{"foo/bar/", "foo/baz"},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var err error

			goldPath, _ := filepath.Abs(filepath.Join("testdata", "golden", strings.Replace(name, " ", "_", -1)))

			tmpdir := chdirToTempdir(t)
			fmt.Println(tmpdir)
			for _, path := range tc.existingPaths {
				if strings.HasSuffix(path, "/") {
					err = os.MkdirAll(path, 0755)
					require.NoError(t, err, "can't create directories")
				} else {
					err = os.MkdirAll(filepath.Dir(path), 0755)
					require.NoError(t, err, "can't create directory for file")

					err = os.WriteFile(path, []byte("some content"), 0644)
					require.NoError(t, err, "could not write sample file")
				}
			}
			m, _ := interactive.InitialModelForTests().Update(nil)

			for _, e := range tc.events {
				m = updateModel(t, m, e)
			}
			out := m.View()
			if stdout {
				fmt.Println(out)
			}

			// Update golden file
			if update {
				t.Logf("updating golden file %s", goldPath)
				err = os.WriteFile(goldPath, []byte(out), 0600)
				require.NoError(t, err, "Cannot write golden file")
			}
			want, err := os.ReadFile(goldPath)
			require.NoError(t, err, "Cannot load policy golden file")

			require.Equal(t, string(want), m.View(), "Didn't get expected output")
		})
	}
}

// updateModel calls Update() on the model and executes returned commands.
// It will reexecute Update() until there are no more returned commands.
func updateModel(t *testing.T, m tea.Model, msg tea.Msg) tea.Model {
	t.Helper()

	m, cmd := m.Update(msg)
	if cmd == nil {
		return m
	}

	messageCandidates := cmd()

	batchMsgType := reflect.TypeOf(tea.Batch(func() tea.Msg { return tea.Msg(struct{}{}) })())

	// executes all messages on batched messages, which is a slice underlying it.
	if reflect.TypeOf(messageCandidates) == batchMsgType {
		if reflect.TypeOf(messageCandidates).Kind() != reflect.Slice {
			t.Fatalf("expected batched messages to be a slice but it's not: %v", reflect.TypeOf(messageCandidates).Kind())
		}

		v := reflect.ValueOf(messageCandidates)
		for i := 0; i < v.Len(); i++ {
			messages := v.Index(i).Call(nil)
			// Call update on all returned messages, which can itself reenter Update()
			for _, msgValue := range messages {
				// if this is a Tick message, ignore it (to avoid endless loop as we will always have the next tick available)
				// and our function is reentrant, not a queue of message. Thus, install is never called.
				if _, ok := msgValue.Interface().(spinner.TickMsg); ok {
					continue
				}

				msg, ok := msgValue.Interface().(tea.Msg)
				if !ok {
					t.Fatalf("expected message to be a tea.Msg, but got: %T", msg)
				}
				m = updateModel(t, m, msg)
			}
		}
		return m
	}

	// We only got one message, call Update() on it
	return updateModel(t, m, messageCandidates)
}

func TestMain(m *testing.M) {
	flag.BoolVar(&update, "update", false, "update golden files")
	flag.BoolVar(&stdout, "stdout", false, "print output to stdout for debugging purposes")
	flag.Parse()

	m.Run()
}

func chdirToTempdir(t *testing.T) string {
	t.Helper()

	orig, err := os.Getwd()
	require.NoError(t, err, "Setup: can't get current directory")

	dir := t.TempDir()
	err = os.Chdir(dir)
	require.NoError(t, err, "Setup: can't change current directory")
	t.Cleanup(func() {
		err := os.Chdir(orig)
		require.NoError(t, err, "Teardown: can't restore current directory")
	})
	return dir
}
