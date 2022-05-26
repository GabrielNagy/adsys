package interactive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ubuntu/adsys/internal/i18n"
	"github.com/ubuntu/adsys/internal/watchdservice"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#99cc99"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	hintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
	boldStyle    = lipgloss.NewStyle().Bold(true)
	titleStyle   = lipgloss.NewStyle().Underline(true).Bold(true)
	focusedStyle = boldStyle.Copy().Foreground(lipgloss.Color("#E95420")) // Ubuntu orange

	submitText    = i18n.G("Install")
	focusedButton = focusedStyle.Copy().Render(fmt.Sprintf("[ %s ]", submitText))
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render(submitText))

	// Add a thick border to the top and bottom.
	border = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true, false)
)

type model struct {
	focusIndex    int
	inputs        []textinput.Model
	spinner       spinner.Model
	defaultConfig string

	err       error
	loading   bool
	typing    bool
	installed bool

	dryrun bool
}

type appConfig struct {
	Verbose int
	Dirs    []string
}

type installMsg struct {
	err error
}

// TODO: check if entered directories actually exist (and show to user)
// TODO: handle existing config file (maybe ask user if they want to overwrite?)
// TODO: golden file testing

// writeConfig writes the config to the given file, checking whether the directories
// that are passed in actually exist.
func (m model) writeConfig(confFile string, dirs []string) error {
	if len(dirs) == 1 && dirs[0] == "" {
		return fmt.Errorf(i18n.G("needs at least one directory to watch"))
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf(i18n.G("directory %q does not exist"), dir)
		}
	}

	// Empty input means using the default config file
	if confFile == "" {
		confFile = m.defaultConfig
	}

	if err := os.MkdirAll(filepath.Dir(confFile), 0755); err != nil {
		return fmt.Errorf("unable to create config directory: %v", err)
	}

	cfg := appConfig{Dirs: dirs, Verbose: 3}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("unable to marshal config: %v", err)
	}

	if err := os.WriteFile(confFile, data, 0644); err != nil {
		return fmt.Errorf("unable to write config: %v", err)
	}

	return nil
}

// installService writes the configuration file and installs the service with
// the file as an argument.
func (m model) installService(confFile string, dirs []string) tea.Cmd {
	return func() tea.Msg {
		// If the user typed in a directory, create the config file inside it
		if stat, err := os.Stat(confFile); err == nil && stat.IsDir() {
			confFile = filepath.Join(confFile, "adwatchd.yml")
		}

		if err := m.writeConfig(confFile, dirs); err != nil {
			return installMsg{err}
		}

		configAbsPath, err := filepath.Abs(confFile)
		if err != nil {
			return installMsg{err}
		}

		svc, err := watchdservice.New(
			context.Background(),
			watchdservice.WithArgs([]string{"-c", configAbsPath}),
		)
		if err != nil {
			return installMsg{err}
		}

		// Only install service on real system
		if m.dryrun {
			return installMsg{nil}
		}

		err = svc.Install(context.Background())
		return installMsg{err}
	}
}

func initialModel(defaultConfig string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	m := model{
		// start with a size of 2 (one for the config path, one for the first
		// configured directory, the slice will be resized based on user input)
		inputs:        make([]textinput.Model, 2),
		spinner:       s,
		typing:        true,
		defaultConfig: defaultConfig,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = newStyledTextInput()

		switch i {
		case 0:
			t.Placeholder = fmt.Sprintf("Config file location (leave blank for default: %s)", m.defaultConfig)
			t.Prompt = "Config file: "
			t.PromptStyle = boldStyle
			t.Focus()
		case 1:
			t.Placeholder = "Directory to watch (one per line)"
		}

		m.inputs[i] = t
	}

	return m
}

func newStyledTextInput() textinput.Model {
	t := textinput.New()
	t.CursorStyle = cursorStyle
	t.CharLimit = 1024
	t.SetCursorMode(textinput.CursorStatic)
	return t
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := teaMsg.(type) {
	case installMsg:
		m.loading = false
		m.installed = true
		if err := msg.err; err != nil {
			m.err = err
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp, tea.KeyShiftTab:
			// Set focus to previous input
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

		case tea.KeyDown, tea.KeyTab:
			// Set focus to next input
			m.focusIndex++
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}

		case tea.KeyBackspace:
			// backspace: set focus to previous input if needed

			// No backspace on config
			if m.focusIndex == 0 {
				break
			}

			// Backspace on submit: go to previous one
			if m.focusIndex == len(m.inputs) {
				m.focusIndex--
				break
			}

			// If element is not empty, let the input widget handling it
			if m.inputs[m.focusIndex].Value() != "" {
				break
			}

			// Handle element removal on any empty directory input
			if m.focusIndex > 1 {
				m.inputs = slices.Delete(m.inputs, m.focusIndex, m.focusIndex+1)
				m.focusIndex--
				// tell that we already handled backspace by changing the message type to nothing
				// Tis prevents input to handle again backspace.
				teaMsg = struct{}{}
				break
			}
			m.focusIndex--

		case tea.KeyEnter:
			// Did the user press enter while the submit button was focused?
			if m.focusIndex == len(m.inputs) {
				var dirs []string
				var confFile string

				for _, i := range m.inputs[1:] {
					if i.Value() != "" {
						dirs = append(dirs, i.Value())
					}
				}

				confFile = m.inputs[0].Value()

				m.typing = false
				m.loading = true

				//return m, m.installService(confFile, dirs)
				return m, tea.Batch(m.spinner.Tick, m.installService(confFile, dirs))
			}

			// Always go to directory from config
			if m.focusIndex == 0 {
				m.focusIndex++
				break
			}

			// Directory fields
			switch m.inputs[m.focusIndex].Value() {
			case "":
				// We need at least one directory to watch. Block action.
				if m.focusIndex == 1 {
					break
				}

				// delete the current (empty) one, focus stays the same index to move to next element
				m.inputs = slices.Delete(m.inputs, m.focusIndex, m.focusIndex+1)

			default:
				if m.inputs[m.focusIndex].Err != nil {
					break
				}
				// add a new input where we are and move focus to it
				m.focusIndex++
				m.inputs = slices.Insert(m.inputs, m.focusIndex, newStyledTextInput())
			}
		}
	}

	// General properties
	if m.installed {
		time.Sleep(time.Second * 2)

		return m, tea.Quit
	}

	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(teaMsg)
		return m, cmd
	}

	if m.typing {
		// Handle character input and blinking
		cmd := m.updateInputs(teaMsg)
		return m, cmd
	}

	return m, nil
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	for i := range m.inputs {
		// Style the input depending on focus
		if i != m.focusIndex {
			// Ensure focused state is removed
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = boldStyle
			m.inputs[i].TextStyle = noStyle
			continue
		}

		// Set focused state
		m.inputs[i].PromptStyle = focusedStyle

		// Record change of focus if current element was not already focused
		if !m.inputs[i].Focused() {
			cmds = append(cmds, m.inputs[i].Focus())
		}

		// Only text inputs with Focus() set will respond, so it's safe to simply
		// update all of them here without any further logic
		var update tea.Cmd
		m.inputs[i], update = m.inputs[i].Update(msg)

		// Update input style/error separately for config and directories
		if m.focusIndex > 0 {
			m.updateDirInputErrorAndStyle(i)
		} else {
			m.updateConfigInputError()
		}
		cmds = append(cmds, update)
	}

	return tea.Batch(cmds...)
}

func (m *model) updateConfigInputError() {
	// If the config input is empty, clean up the error message
	if m.inputs[0].Value() == "" {
		m.inputs[0].Err = nil
		return
	}

	stat, err := os.Stat(m.inputs[0].Value())

	// If the config file does not exist, we're good
	if errors.Is(err, os.ErrNotExist) {
		m.inputs[0].Err = nil
		return
	}

	// If we got another error, display it
	if err != nil {
		m.inputs[0].Err = err
		return
	}

	if stat.IsDir() {
		m.inputs[0].Err = errors.New("is a directory; will create adwatchd.yml inside")
		return
	}

	if stat.Mode().IsRegular() {
		m.inputs[0].Err = errors.New("file already exists and will be overwritten")
		return
	}

	m.inputs[0].Err = nil
}

func (m *model) updateDirInputErrorAndStyle(i int) {
	// We consider an empty string to be valid, so users are allowed to press
	// enter on it.
	if m.inputs[i].Value() == "" {
		m.inputs[i].Err = nil
		return
	}

	// Check to see if the directory exists
	if stat, err := os.Stat(m.inputs[i].Value()); errors.Is(err, os.ErrNotExist) || !stat.IsDir() {
		m.inputs[i].Err = errors.New("directory does not exist, please enter a valid path")
		m.inputs[i].TextStyle = noStyle
	} else {
		m.inputs[i].Err = nil
		m.inputs[i].TextStyle = successStyle
	}
}

func (m model) View() string {
	if m.loading {
		return fmt.Sprintf("%s installing service... please wait.", m.spinner.View())
	}

	if err := m.err; err != nil {
		return fmt.Sprintf("Could not install service: %v\n", err)
	}

	if m.typing {
		var b strings.Builder

		b.WriteString(titleStyle.Render("Ubuntu AD Watch Daemon Installer"))
		b.WriteString("\n\n")

		// Display config input and hint
		b.WriteString(m.inputs[0].View())
		if m.inputs[0].Err != nil {
			b.WriteRune('\n')
			b.WriteString(hintStyle.Render(fmt.Sprintf("%s: %s", m.inputs[0].Value(), m.inputs[0].Err.Error())))
			b.WriteString("\n\n")
		} else {
			b.WriteString("\n\n\n")
		}

		if m.focusIndex > 0 && m.focusIndex < len(m.inputs) {
			b.WriteString(focusedStyle.Render("Directories:"))
		} else {
			b.WriteString(boldStyle.Render("Directories:"))
		}
		b.WriteRune('\n')

		// Display directory inputs
		for i, v := range m.inputs[1:] {
			_, _ = b.WriteString(v.View())
			if i < len(m.inputs)-1 {
				_, _ = b.WriteRune('\n')
			}
		}

		// Display directory error if any
		if m.focusIndex > 0 && m.focusIndex < len(m.inputs) && m.inputs[m.focusIndex].Err != nil {
			b.WriteString(hintStyle.Render(fmt.Sprintf("%s: %s", m.inputs[m.focusIndex].Value(), m.inputs[m.focusIndex].Err.Error())))
		}

		// Display button
		button := &blurredButton
		if m.focusIndex == len(m.inputs) {
			button = &focusedButton
		}
		_, _ = fmt.Fprintf(&b, "\n\n%s\n", *button)

		return b.String()
	}

	return fmt.Sprintln("Service adwatchd was successfully installed and is now running.")
}

// Start starts the interactive user experience.
func Start(ctx context.Context, defaultConfig string) error {
	p := tea.NewProgram(initialModel(defaultConfig))
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}
