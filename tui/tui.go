package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tom/shippost/config"
	"github.com/tom/shippost/git"
	"github.com/tom/shippost/x"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)
)

type state int

const (
	stateSelectCommit state = iota
	stateCompose
	statePosting
	statePosted
)

// Model is the main TUI model
type Model struct {
	commits      []git.Commit
	cursor       int
	state        state
	textarea     textarea.Model
	viewport     viewport.Model
	status       string
	err          error
	postURL      string
	width        int
	height       int
	xClient      *x.Client
	cfg          *config.Config
}

// PostResult is returned when posting completes
type postResultMsg struct {
	url string
	err error
}

// New creates a new TUI model
func New() (Model, error) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return Model{}, err
	}

	// Get recent commits
	commits, err := git.GetRecentCommits(10)
	if err != nil {
		return Model{}, err
	}

	// Create textarea
	ta := textarea.New()
	ta.Placeholder = "Write your post..."
	ta.CharLimit = 280
	ta.SetWidth(60)
	ta.SetHeight(4)

	return Model{
		commits:  commits,
		cursor:   0,
		state:    stateSelectCommit,
		textarea: ta,
		xClient:  x.NewClient(cfg),
		cfg:      cfg,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateSelectCommit:
			switch msg.String() {
			case "q", "ctrl+c", "esc":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.commits)-1 {
					m.cursor++
				}
			case "enter":
				// Move to compose with selected commit
				m.state = stateCompose
				commit := m.commits[m.cursor]
				m.textarea.SetValue(commit.Subject)
				m.textarea.Focus()
				return m, textarea.Blink
			case "n":
				// New post without commit context
				m.state = stateCompose
				m.textarea.SetValue("")
				m.textarea.Focus()
				return m, textarea.Blink
			}

		case stateCompose:
			switch msg.String() {
			case "esc":
				m.state = stateSelectCommit
				m.textarea.Blur()
				return m, nil
			case "ctrl+p":
				// Post
				if strings.TrimSpace(m.textarea.Value()) == "" {
					m.err = fmt.Errorf("post cannot be empty")
					return m, nil
				}
				m.state = statePosting
				m.status = "Posting..."
				return m, m.doPost()
			case "ctrl+c":
				return m, tea.Quit
			}
			// Update textarea
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)

		case statePosted:
			switch msg.String() {
			case "q", "ctrl+c", "esc", "enter":
				return m, tea.Quit
			case "n":
				// New post
				m.state = stateSelectCommit
				m.status = ""
				m.postURL = ""
				m.err = nil
				return m, nil
			}
		}

	case postResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateCompose
			m.status = ""
		} else {
			m.state = statePosted
			m.postURL = msg.url
			m.status = "Posted successfully!"
			m.err = nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(min(60, msg.Width-4))
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("shippost"))
	b.WriteString("\n\n")

	switch m.state {
	case stateSelectCommit:
		b.WriteString(dimStyle.Render("Recent commits:"))
		b.WriteString("\n\n")

		for i, commit := range m.commits {
			cursor := "  "
			style := normalStyle
			if i == m.cursor {
				cursor = "● "
				style = selectedStyle
			}

			line := fmt.Sprintf("%s%s %s",
				cursor,
				dimStyle.Render(commit.Ago),
				style.Render(truncate(commit.Subject, 50)),
			)
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString(helpStyle.Render("\n↑/↓ navigate • enter select • n new post • q quit"))

	case stateCompose, statePosting:
		if m.cursor < len(m.commits) {
			commit := m.commits[m.cursor]
			b.WriteString(dimStyle.Render(fmt.Sprintf("Based on: %s", commit.Subject)))
			b.WriteString("\n\n")
		}

		b.WriteString(boxStyle.Render(m.textarea.View()))
		b.WriteString("\n")

		charCount := len(m.textarea.Value())
		countStyle := dimStyle
		if charCount > 260 {
			countStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		}
		if charCount > 280 {
			countStyle = errorStyle
		}
		b.WriteString(countStyle.Render(fmt.Sprintf("%d/280", charCount)))

		if m.err != nil {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render(m.err.Error()))
		}

		if m.state == statePosting {
			b.WriteString("\n")
			b.WriteString(statusStyle.Render(m.status))
		}

		b.WriteString(helpStyle.Render("\nctrl+p post • esc back • ctrl+c quit"))

	case statePosted:
		b.WriteString(statusStyle.Render("✓ " + m.status))
		b.WriteString("\n\n")
		b.WriteString(m.postURL)
		b.WriteString(helpStyle.Render("\n\nn new post • q quit"))
	}

	return b.String()
}

func (m Model) doPost() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.xClient.Post(m.textarea.Value())
		if err != nil {
			return postResultMsg{err: err}
		}
		return postResultMsg{url: fmt.Sprintf("https://x.com/i/status/%s", resp.Data.ID)}
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Run starts the TUI
func Run() error {
	m, err := New()
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
