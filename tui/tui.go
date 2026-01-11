package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tom/shippost/ai"
	"github.com/tom/shippost/config"
	"github.com/tom/shippost/git"
	"github.com/tom/shippost/x"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B"))

	taglineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0")).
			Bold(true)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))

	menuDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B")).
			PaddingLeft(4)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFE66D")).
			Bold(true)

	selectedDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				PaddingLeft(4)

	bulletStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	dimBulletStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#475569"))

	disabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#475569"))

	disabledTagStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#334155")).
				Background(lipgloss.Color("#1E293B")).
				Padding(0, 1)

	disabledDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#334155")).
				PaddingLeft(4)

	helpBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B")).
			Border(lipgloss.Border{Top: "─"}).
			BorderForeground(lipgloss.Color("#334155")).
			PaddingTop(1).
			MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0EA5E9")).
			Background(lipgloss.Color("#0C4A6E")).
			Padding(0, 1).
			Bold(true)

	helpTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(0, 1)

	activeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4ECDC4")).
			Padding(0, 1)

	urlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			Underline(true)

	mediaTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Background(lipgloss.Color("#064E3B")).
			Padding(0, 1)

	threadNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6366F1")).
			Background(lipgloss.Color("#312E81")).
			Padding(0, 1).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8"))

	commitHashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95E6CB"))

	commitTimeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA"))

	aiTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F472B6")).
			Background(lipgloss.Color("#831843")).
			Padding(0, 1).
			Bold(true)
)

type state int

const (
	stateHome state = iota
	stateCompose
	stateMediaInput
	statePosting
	statePosted
	stateCommitBrowser
	stateGenerating
	stateSmartCompose
)

type menuItem struct {
	title       string
	description string
	enabled     bool
}

type threadItem struct {
	text     string
	mediaIDs []string
	media    []string
}

// Model is the main TUI model
type Model struct {
	state           state
	menuCursor      int
	menuItems       []menuItem
	textarea        textarea.Model
	pathInput       textinput.Model
	thread          []threadItem
	currentPost     int
	status          string
	err             error
	postURL         string
	postURLs        []string
	width           int
	height          int
	xClient         *x.Client
	cfg             *config.Config
	commits         []git.Commit
	commitCursor    int
	selectedCommits []int
	aiSuggestion    string
	isSmartPost     bool
}

// Messages
type postResultMsg struct {
	urls []string
	err  error
}

type mediaUploadMsg struct {
	mediaID string
	path    string
	err     error
}

type commitsLoadedMsg struct {
	commits []git.Commit
	err     error
}

type aiSuggestionMsg struct {
	suggestion string
	err        error
}

// New creates a new TUI model
func New() (Model, error) {
	cfg, err := config.Load()
	if err != nil {
		return Model{}, err
	}

	ta := textarea.New()
	ta.Placeholder = "What's happening?"
	ta.CharLimit = 280
	ta.SetWidth(60)
	ta.SetHeight(5)
	ta.ShowLineNumbers = false

	pi := textinput.New()
	pi.Placeholder = "~/Pictures/screenshot.png"
	pi.Width = 50
	pi.CharLimit = 256

	claudeAvailable := ai.IsClaudeAvailable()
	smartPostDesc := "AI-powered posts to X from your commits"
	if !claudeAvailable {
		smartPostDesc = "Requires Claude Code CLI (not installed)"
	}

	menuItems := []menuItem{
		{
			title:       "Quick Post",
			description: "Write and post to X now",
			enabled:     true,
		},
		{
			title:       "Smart Post",
			description: smartPostDesc,
			enabled:     claudeAvailable,
		},
	}

	return Model{
		state:           stateHome,
		menuCursor:      0,
		menuItems:       menuItems,
		textarea:        ta,
		pathInput:       pi,
		thread:          []threadItem{{text: "", mediaIDs: nil, media: nil}},
		currentPost:     0,
		xClient:         x.NewClient(cfg),
		cfg:             cfg,
		commits:         nil,
		commitCursor:    0,
		selectedCommits: nil,
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
		case stateHome:
			switch msg.String() {
			case "q", "ctrl+c", "esc":
				return m, tea.Quit
			case "up", "k":
				if m.menuCursor > 0 {
					m.menuCursor--
				}
			case "down", "j":
				if m.menuCursor < len(m.menuItems)-1 {
					m.menuCursor++
				}
			case "enter":
				item := m.menuItems[m.menuCursor]
				if item.enabled {
					if m.menuCursor == 0 {
						// Quick Post
						m.state = stateCompose
						m.isSmartPost = false
						m.thread = []threadItem{{text: "", mediaIDs: nil, media: nil}}
						m.currentPost = 0
						m.textarea.SetValue("")
						m.textarea.Focus()
						return m, textarea.Blink
					} else if m.menuCursor == 1 {
						// Smart Post - load commits
						m.state = stateCommitBrowser
						m.isSmartPost = true
						m.commitCursor = 0
						m.selectedCommits = nil
						m.status = "Loading commits..."
						return m, m.loadCommits()
					}
				}
			}

		case stateCommitBrowser:
			switch msg.String() {
			case "esc", "q":
				m.state = stateHome
				m.commits = nil
				m.selectedCommits = nil
				m.err = nil
				return m, nil
			case "up", "k":
				if m.commitCursor > 0 {
					m.commitCursor--
				}
			case "down", "j":
				if m.commitCursor < len(m.commits)-1 {
					m.commitCursor++
				}
			case " ":
				// Toggle selection
				idx := m.commitCursor
				found := false
				for i, s := range m.selectedCommits {
					if s == idx {
						m.selectedCommits = append(m.selectedCommits[:i], m.selectedCommits[i+1:]...)
						found = true
						break
					}
				}
				if !found {
					m.selectedCommits = append(m.selectedCommits, idx)
				}
			case "enter":
				// Generate AI suggestion
				if len(m.selectedCommits) == 0 {
					m.selectedCommits = []int{m.commitCursor}
				}
				m.state = stateGenerating
				m.status = "Generating suggestion..."
				return m, m.generateSuggestion()
			case "ctrl+c":
				return m, tea.Quit
			}

		case stateSmartCompose:
			switch msg.String() {
			case "esc":
				m.state = stateCommitBrowser
				m.textarea.Blur()
				m.err = nil
				return m, nil
			case "ctrl+s":
				m.thread[0].text = m.textarea.Value()
				if !m.hasContent() {
					m.err = fmt.Errorf("post cannot be empty")
					return m, nil
				}
				m.state = statePosting
				m.status = "Posting..."
				return m, m.doPost()
			case "ctrl+r":
				// Regenerate suggestion
				m.state = stateGenerating
				m.status = "Regenerating..."
				return m, m.generateSuggestion()
			case "ctrl+o":
				m.thread[0].text = m.textarea.Value()
				m.state = stateMediaInput
				m.pathInput.SetValue("")
				m.pathInput.Focus()
				return m, textinput.Blink
			case "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)

		case stateCompose:
			switch msg.String() {
			case "esc":
				m.state = stateHome
				m.textarea.Blur()
				m.err = nil
				m.thread = []threadItem{{text: "", mediaIDs: nil, media: nil}}
				m.currentPost = 0
				return m, nil
			case "ctrl+s":
				m.thread[m.currentPost].text = m.textarea.Value()
				if !m.hasContent() {
					m.err = fmt.Errorf("post cannot be empty")
					return m, nil
				}
				m.state = statePosting
				m.status = "Posting..."
				return m, m.doPost()
			case "ctrl+o":
				m.thread[m.currentPost].text = m.textarea.Value()
				m.state = stateMediaInput
				m.pathInput.SetValue("")
				m.pathInput.Focus()
				return m, textinput.Blink
			case "ctrl+n":
				m.thread[m.currentPost].text = m.textarea.Value()
				m.thread = append(m.thread, threadItem{text: "", mediaIDs: nil, media: nil})
				m.currentPost = len(m.thread) - 1
				m.textarea.SetValue("")
				return m, nil
			case "ctrl+up", "ctrl+k":
				if m.currentPost > 0 {
					m.thread[m.currentPost].text = m.textarea.Value()
					m.currentPost--
					m.textarea.SetValue(m.thread[m.currentPost].text)
				}
				return m, nil
			case "ctrl+down", "ctrl+j":
				if m.currentPost < len(m.thread)-1 {
					m.thread[m.currentPost].text = m.textarea.Value()
					m.currentPost++
					m.textarea.SetValue(m.thread[m.currentPost].text)
				}
				return m, nil
			case "ctrl+x":
				if len(m.thread[m.currentPost].media) > 0 {
					m.thread[m.currentPost].media = m.thread[m.currentPost].media[:len(m.thread[m.currentPost].media)-1]
					m.thread[m.currentPost].mediaIDs = m.thread[m.currentPost].mediaIDs[:len(m.thread[m.currentPost].mediaIDs)-1]
				}
				return m, nil
			case "ctrl+d":
				if len(m.thread) > 1 {
					m.thread = append(m.thread[:m.currentPost], m.thread[m.currentPost+1:]...)
					if m.currentPost >= len(m.thread) {
						m.currentPost = len(m.thread) - 1
					}
					m.textarea.SetValue(m.thread[m.currentPost].text)
				}
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)

		case stateMediaInput:
			switch msg.String() {
			case "esc":
				if m.isSmartPost {
					m.state = stateSmartCompose
				} else {
					m.state = stateCompose
				}
				m.pathInput.Blur()
				m.textarea.Focus()
				return m, textarea.Blink
			case "enter":
				path := m.pathInput.Value()
				if path == "" {
					if m.isSmartPost {
						m.state = stateSmartCompose
					} else {
						m.state = stateCompose
					}
					m.textarea.Focus()
					return m, textarea.Blink
				}
				if strings.HasPrefix(path, "~/") {
					home, _ := os.UserHomeDir()
					path = filepath.Join(home, path[2:])
				}
				ext := strings.ToLower(filepath.Ext(path))
				validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
				if !validExts[ext] {
					m.err = fmt.Errorf("unsupported file type: %s", ext)
					if m.isSmartPost {
						m.state = stateSmartCompose
					} else {
						m.state = stateCompose
					}
					m.textarea.Focus()
					return m, textarea.Blink
				}
				m.status = "Uploading..."
				return m, m.uploadMedia(path)
			case "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.pathInput, cmd = m.pathInput.Update(msg)
			cmds = append(cmds, cmd)

		case statePosted:
			switch msg.String() {
			case "q", "ctrl+c", "esc", "enter":
				return m, tea.Quit
			case "n":
				m.state = stateHome
				m.status = ""
				m.postURL = ""
				m.postURLs = nil
				m.err = nil
				m.thread = []threadItem{{text: "", mediaIDs: nil, media: nil}}
				m.currentPost = 0
				m.isSmartPost = false
				return m, nil
			}
		}

	case commitsLoadedMsg:
		m.status = ""
		if msg.err != nil {
			m.err = msg.err
			m.state = stateHome
		} else {
			m.commits = msg.commits
		}

	case aiSuggestionMsg:
		m.status = ""
		if msg.err != nil {
			m.err = msg.err
			m.state = stateCommitBrowser
		} else {
			m.aiSuggestion = msg.suggestion
			m.state = stateSmartCompose
			m.thread = []threadItem{{text: msg.suggestion, mediaIDs: nil, media: nil}}
			m.currentPost = 0
			m.textarea.SetValue(msg.suggestion)
			m.textarea.Focus()
			return m, textarea.Blink
		}

	case mediaUploadMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = ""
		} else {
			if m.isSmartPost {
				m.thread[0].mediaIDs = append(m.thread[0].mediaIDs, msg.mediaID)
				m.thread[0].media = append(m.thread[0].media, msg.path)
			} else {
				m.thread[m.currentPost].mediaIDs = append(m.thread[m.currentPost].mediaIDs, msg.mediaID)
				m.thread[m.currentPost].media = append(m.thread[m.currentPost].media, msg.path)
			}
			m.status = ""
		}
		if m.isSmartPost {
			m.state = stateSmartCompose
		} else {
			m.state = stateCompose
		}
		m.textarea.Focus()
		return m, textarea.Blink

	case postResultMsg:
		if msg.err != nil {
			m.err = msg.err
			if m.isSmartPost {
				m.state = stateSmartCompose
			} else {
				m.state = stateCompose
			}
			m.status = ""
		} else {
			m.state = statePosted
			m.postURLs = msg.urls
			if len(msg.urls) > 0 {
				m.postURL = msg.urls[0]
			}
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
	b.WriteString("  ")
	b.WriteString(taglineStyle.Render("Share your work with the world"))
	b.WriteString("\n\n")

	switch m.state {
	case stateHome:
		for i, item := range m.menuItems {
			if i == m.menuCursor {
				b.WriteString(bulletStyle.Render("▸ "))
				if item.enabled {
					b.WriteString(selectedStyle.Render(item.title))
				} else {
					b.WriteString(disabledStyle.Render(item.title))
				}
				b.WriteString("\n")
				if item.enabled {
					b.WriteString(selectedDescStyle.Render(item.description))
				} else {
					b.WriteString(disabledDescStyle.Render(item.description))
				}
			} else {
				b.WriteString(dimBulletStyle.Render("  "))
				if item.enabled {
					b.WriteString(menuItemStyle.Render(item.title))
				} else {
					b.WriteString(disabledStyle.Render(item.title))
				}
				b.WriteString("\n")
				if item.enabled {
					b.WriteString(menuDescStyle.Render(item.description))
				} else {
					b.WriteString(disabledDescStyle.Render(item.description))
				}
			}
			b.WriteString("\n\n")
		}

		b.WriteString(m.renderHelpBar([]helpItem{
			{"↑↓", "navigate"},
			{"enter", "select"},
			{"q", "quit"},
		}))

	case stateCommitBrowser:
		b.WriteString(subtitleStyle.Render("Smart Post"))
		b.WriteString("  ")
		b.WriteString(dimStyle.Render("Select commits to post about"))
		b.WriteString("\n\n")

		if m.status != "" {
			b.WriteString(statusStyle.Render("● " + m.status))
			b.WriteString("\n")
		} else if m.err != nil {
			b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
			b.WriteString("\n")
		} else if len(m.commits) == 0 {
			b.WriteString(dimStyle.Render("No commits found in this repository"))
		} else {
			for i, commit := range m.commits {
				isSelected := false
				for _, s := range m.selectedCommits {
					if s == i {
						isSelected = true
						break
					}
				}

				if i == m.commitCursor {
					b.WriteString(bulletStyle.Render("▸ "))
				} else {
					b.WriteString("  ")
				}

				if isSelected {
					b.WriteString(selectedStyle.Render("● "))
				} else {
					b.WriteString(dimStyle.Render("○ "))
				}

				b.WriteString(commitTimeStyle.Render(fmt.Sprintf("%-12s ", commit.Ago)))

				if i == m.commitCursor {
					b.WriteString(selectedStyle.Render(truncate(commit.Subject, 45)))
				} else {
					b.WriteString(menuItemStyle.Render(truncate(commit.Subject, 45)))
				}
				b.WriteString("\n")
			}

			if len(m.selectedCommits) > 0 {
				b.WriteString("\n")
				b.WriteString(dimStyle.Render(fmt.Sprintf("%d commit(s) selected", len(m.selectedCommits))))
			}
		}

		b.WriteString("\n")
		b.WriteString(m.renderHelpBar([]helpItem{
			{"↑↓", "navigate"},
			{"space", "select"},
			{"enter", "generate"},
			{"esc", "back"},
		}))

	case stateGenerating:
		b.WriteString(subtitleStyle.Render("Smart Post"))
		b.WriteString("\n\n")
		b.WriteString(statusStyle.Render("● " + m.status))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("Claude is writing your post..."))

	case stateSmartCompose:
		b.WriteString(subtitleStyle.Render("Smart Post"))
		b.WriteString("  ")
		b.WriteString(aiTagStyle.Render(" AI "))
		b.WriteString("\n\n")

		b.WriteString(activeBoxStyle.Render(m.textarea.View()))
		b.WriteString("\n")

		charCount := len(m.textarea.Value())
		countStyle := helpTextStyle
		if charCount > 260 {
			countStyle = warningStyle
		}
		if charCount > 280 {
			countStyle = errorStyle
		}
		b.WriteString(countStyle.Render(fmt.Sprintf("%d", charCount)))
		b.WriteString(helpTextStyle.Render("/280"))

		if len(m.thread[0].media) > 0 {
			b.WriteString("  ")
			for _, path := range m.thread[0].media {
				b.WriteString(mediaTagStyle.Render(" 📎 " + filepath.Base(path) + " "))
				b.WriteString(" ")
			}
		}

		if m.err != nil {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		}

		b.WriteString("\n")
		b.WriteString(m.renderHelpBar([]helpItem{
			{"ctrl+s", "send"},
			{"ctrl+r", "regenerate"},
			{"ctrl+o", "attach"},
			{"esc", "back"},
		}))

	case stateCompose, statePosting:
		if len(m.thread) > 1 {
			b.WriteString(subtitleStyle.Render("Quick Post"))
			b.WriteString("  ")
			b.WriteString(threadNumStyle.Render(fmt.Sprintf(" %d/%d ", m.currentPost+1, len(m.thread))))
		} else {
			b.WriteString(subtitleStyle.Render("Quick Post"))
		}
		b.WriteString("\n\n")

		if m.state == stateCompose {
			b.WriteString(activeBoxStyle.Render(m.textarea.View()))
		} else {
			b.WriteString(boxStyle.Render(m.textarea.View()))
		}
		b.WriteString("\n")

		charCount := len(m.textarea.Value())
		countStyle := helpTextStyle
		if charCount > 260 {
			countStyle = warningStyle
		}
		if charCount > 280 {
			countStyle = errorStyle
		}
		b.WriteString(countStyle.Render(fmt.Sprintf("%d", charCount)))
		b.WriteString(helpTextStyle.Render("/280"))

		if len(m.thread[m.currentPost].media) > 0 {
			b.WriteString("  ")
			for _, path := range m.thread[m.currentPost].media {
				b.WriteString(mediaTagStyle.Render(" 📎 " + filepath.Base(path) + " "))
				b.WriteString(" ")
			}
		}

		if m.err != nil {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		}

		if m.state == statePosting {
			b.WriteString("\n")
			b.WriteString(statusStyle.Render("● " + m.status))
		}

		if len(m.thread) > 1 {
			b.WriteString("\n\n")
			b.WriteString(dimStyle.Render("Thread:"))
			b.WriteString("\n")
			for i, item := range m.thread {
				prefix := "  "
				if i == m.currentPost {
					prefix = "▸ "
					b.WriteString(selectedStyle.Render(prefix))
				} else {
					b.WriteString(dimStyle.Render(prefix))
				}
				preview := item.text
				if i == m.currentPost {
					preview = m.textarea.Value()
				}
				if len(preview) > 40 {
					preview = preview[:37] + "..."
				}
				if preview == "" {
					preview = "(empty)"
				}
				if i == m.currentPost {
					b.WriteString(selectedStyle.Render(fmt.Sprintf("%d. %s", i+1, preview)))
				} else {
					b.WriteString(dimStyle.Render(fmt.Sprintf("%d. %s", i+1, preview)))
				}
				if len(item.media) > 0 {
					b.WriteString(dimStyle.Render(fmt.Sprintf(" [%d media]", len(item.media))))
				}
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
		helpItems := []helpItem{
			{"ctrl+s", "send"},
			{"ctrl+o", "attach"},
			{"ctrl+n", "add"},
		}
		if len(m.thread) > 1 {
			helpItems = append(helpItems, helpItem{"ctrl+d", "delete"})
			helpItems = append(helpItems, helpItem{"ctrl+↑↓", "nav"})
		}
		helpItems = append(helpItems, helpItem{"esc", "back"})
		b.WriteString(m.renderHelpBar(helpItems))

	case stateMediaInput:
		b.WriteString(subtitleStyle.Render("Attach Image"))
		b.WriteString("\n\n")
		b.WriteString(inputLabelStyle.Render("File path:"))
		b.WriteString("\n")
		b.WriteString(activeBoxStyle.Render(m.pathInput.View()))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Supports: .jpg .png .gif .webp • Use ~/path for home"))

		if m.status != "" {
			b.WriteString("\n")
			b.WriteString(statusStyle.Render("● " + m.status))
		}

		b.WriteString("\n")
		b.WriteString(m.renderHelpBar([]helpItem{
			{"enter", "upload"},
			{"esc", "cancel"},
		}))

	case statePosted:
		if len(m.postURLs) > 1 {
			b.WriteString(statusStyle.Render(fmt.Sprintf("✓ Thread posted! (%d posts)", len(m.postURLs))))
		} else {
			b.WriteString(statusStyle.Render("✓ Posted successfully!"))
		}
		b.WriteString("\n\n")

		for i, url := range m.postURLs {
			if len(m.postURLs) > 1 {
				b.WriteString(dimStyle.Render(fmt.Sprintf("%d. ", i+1)))
			}
			b.WriteString(urlStyle.Render(url))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(m.renderHelpBar([]helpItem{
			{"n", "new post"},
			{"q", "quit"},
		}))
	}

	return b.String()
}

type helpItem struct {
	key  string
	text string
}

func (m Model) renderHelpBar(items []helpItem) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, helpKeyStyle.Render(item.key)+" "+helpTextStyle.Render(item.text))
	}
	return helpBarStyle.Render(strings.Join(parts, "   "))
}

func (m Model) hasContent() bool {
	for _, item := range m.thread {
		if strings.TrimSpace(item.text) != "" || len(item.mediaIDs) > 0 {
			return true
		}
	}
	return strings.TrimSpace(m.textarea.Value()) != ""
}

func (m Model) loadCommits() tea.Cmd {
	return func() tea.Msg {
		commits, err := git.GetRecentCommits(15)
		if err != nil {
			return commitsLoadedMsg{err: err}
		}
		return commitsLoadedMsg{commits: commits}
	}
}

func (m Model) generateSuggestion() tea.Cmd {
	return func() tea.Msg {
		var selectedCommits []git.Commit
		for _, idx := range m.selectedCommits {
			if idx < len(m.commits) {
				selectedCommits = append(selectedCommits, m.commits[idx])
			}
		}

		if !ai.IsClaudeAvailable() {
			return aiSuggestionMsg{err: fmt.Errorf("claude CLI not found - install Claude Code first")}
		}

		suggestion, err := ai.GeneratePostSuggestion(selectedCommits)
		if err != nil {
			return aiSuggestionMsg{err: err}
		}
		return aiSuggestionMsg{suggestion: suggestion}
	}
}

func (m Model) uploadMedia(path string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.xClient.UploadMedia(path)
		if err != nil {
			return mediaUploadMsg{err: err}
		}
		return mediaUploadMsg{mediaID: resp.MediaIDString, path: path}
	}
}

func (m Model) doPost() tea.Cmd {
	return func() tea.Msg {
		var posts []x.ThreadPost
		for i, item := range m.thread {
			text := item.text
			if i == m.currentPost {
				text = m.textarea.Value()
			}
			if strings.TrimSpace(text) == "" && len(item.mediaIDs) == 0 {
				continue
			}
			posts = append(posts, x.ThreadPost{
				Text:     text,
				MediaIDs: item.mediaIDs,
			})
		}

		if len(posts) == 0 {
			return postResultMsg{err: fmt.Errorf("no content to post")}
		}

		if len(posts) == 1 {
			resp, err := m.xClient.PostWithOptions(posts[0].Text, &x.PostOptions{
				MediaIDs: posts[0].MediaIDs,
			})
			if err != nil {
				return postResultMsg{err: err}
			}
			return postResultMsg{urls: []string{fmt.Sprintf("https://x.com/i/status/%s", resp.Data.ID)}}
		}

		responses, err := m.xClient.PostThread(posts)
		if err != nil {
			return postResultMsg{err: err}
		}

		var urls []string
		for _, resp := range responses {
			urls = append(urls, fmt.Sprintf("https://x.com/i/status/%s", resp.Data.ID))
		}
		return postResultMsg{urls: urls}
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
