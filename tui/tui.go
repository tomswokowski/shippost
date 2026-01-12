package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tomswokowski/shippost/ai"
	"github.com/tomswokowski/shippost/config"
	"github.com/tomswokowski/shippost/git"
	"github.com/tomswokowski/shippost/x"
)

type state int

const (
	stateHome state = iota
	stateCompose
	stateMediaInput
	statePosting
	statePosted
	stateSmartMenu
	stateCommitBrowser
	stateAskInput
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
	state              state
	menuCursor         int
	menuItems          []menuItem
	textarea           textarea.Model
	pathInput          textinput.Model
	askInput           textarea.Model
	commitPromptInput  textarea.Model
	thread             []threadItem
	currentPost        int
	status             string
	err                error
	postURL            string
	postURLs           []string
	width              int
	height             int
	xClient            *x.Client
	cfg                *config.Config
	commits            []git.Commit
	commitCursor       int
	selectedCommits    []int
	aiSuggestion       string
	isSmartPost        bool
	smartMenuCursor    int
	askQuery           string
	commitPromptActive bool
	commitScrollOffset int
	commitSearch       string
	commitSearchActive bool
	filteredCommits    []int
	allowThread        bool
	inGitRepo          bool
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

	askIn := textarea.New()
	askIn.Placeholder = "What did I accomplish today?"
	askIn.SetWidth(55)
	askIn.SetHeight(3)
	askIn.CharLimit = 500
	askIn.ShowLineNumbers = false

	commitPrompt := textarea.New()
	commitPrompt.Placeholder = "Optional: focus on performance, make it casual, etc."
	commitPrompt.SetWidth(55)
	commitPrompt.SetHeight(2)
	commitPrompt.CharLimit = 500
	commitPrompt.ShowLineNumbers = false

	claudeAvailable := ai.IsClaudeAvailable()
	inGitRepo := git.IsGitRepo()

	smartPostEnabled := claudeAvailable && inGitRepo
	smartPostDesc := "AI-powered posts from your git commits"
	if !inGitRepo {
		smartPostDesc = "Not in a git repository"
	} else if !claudeAvailable {
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
			enabled:     smartPostEnabled,
		},
	}

	return Model{
		state:             stateHome,
		menuCursor:        0,
		menuItems:         menuItems,
		textarea:          ta,
		pathInput:         pi,
		askInput:          askIn,
		commitPromptInput: commitPrompt,
		thread:            []threadItem{{text: "", mediaIDs: nil, media: nil}},
		currentPost:       0,
		xClient:           x.NewClient(cfg),
		cfg:               cfg,
		commits:           nil,
		commitCursor:      0,
		selectedCommits:   nil,
		allowThread:       true,
		inGitRepo:         inGitRepo,
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
			return m.handleHomeKeys(msg)
		case stateSmartMenu:
			return m.handleSmartMenuKeys(msg)
		case stateAskInput:
			return m.handleAskInputKeys(msg)
		case stateCommitBrowser:
			return m.handleCommitBrowserKeys(msg)
		case stateSmartCompose:
			return m.handleComposeKeys(msg, true)
		case stateCompose:
			return m.handleComposeKeys(msg, false)
		case stateMediaInput:
			return m.handleMediaInputKeys(msg)
		case statePosted:
			return m.handlePostedKeys(msg)
		}

	case commitsLoadedMsg:
		m.status = ""
		if msg.err != nil {
			m.err = msg.err
			m.state = stateHome
		} else {
			m.commits = msg.commits
			m.filteredCommits = make([]int, len(msg.commits))
			for i := range msg.commits {
				m.filteredCommits[i] = i
			}
		}

	case aiSuggestionMsg:
		m.status = ""
		if msg.err != nil {
			m.err = msg.err
			m.state = stateSmartMenu
		} else {
			m.thread = nil
			for _, suggestion := range msg.suggestions {
				m.thread = append(m.thread, threadItem{text: suggestion, mediaIDs: nil, media: nil})
			}
			if len(m.thread) == 0 {
				m.thread = []threadItem{{text: "", mediaIDs: nil, media: nil}}
			}
			m.state = stateSmartCompose
			m.currentPost = 0
			m.textarea.SetValue(m.thread[0].text)
			m.textarea.Focus()
			return m, textarea.Blink
		}

	case mediaUploadMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = ""
		} else {
			m.thread[m.currentPost].mediaIDs = append(m.thread[m.currentPost].mediaIDs, msg.mediaID)
			m.thread[m.currentPost].media = append(m.thread[m.currentPost].media, msg.path)
			m.status = ""
			m.err = nil
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
		return m, tea.ClearScreen
	}

	return m, tea.Batch(cmds...)
}

// Key handlers for each state

func (m Model) handleHomeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
				m.state = stateCompose
				m.isSmartPost = false
				m.thread = []threadItem{{text: "", mediaIDs: nil, media: nil}}
				m.currentPost = 0
				m.textarea.SetValue("")
				m.textarea.Focus()
				return m, textarea.Blink
			} else if m.menuCursor == 1 {
				m.state = stateSmartMenu
				m.isSmartPost = true
				m.smartMenuCursor = 0
				m.err = nil
				return m, nil
			}
		}
	}
	return m, nil
}

func (m Model) handleSmartMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = stateHome
		m.err = nil
		return m, nil
	case "up", "k":
		if m.smartMenuCursor > 0 {
			m.smartMenuCursor--
		}
	case "down", "j":
		if m.smartMenuCursor < 1 {
			m.smartMenuCursor++
		}
	case "enter":
		if m.smartMenuCursor == 0 {
			m.state = stateCommitBrowser
			m.commitCursor = 0
			m.selectedCommits = nil
			m.commitPromptInput.SetValue("")
			m.commitPromptActive = false
			m.commitScrollOffset = 0
			m.commitSearch = ""
			m.commitSearchActive = false
			m.filteredCommits = nil
			m.status = "Loading commits..."
			return m, m.loadCommits()
		} else {
			m.state = stateAskInput
			m.askInput.SetValue("")
			m.askInput.Focus()
			m.status = "Loading commits..."
			return m, tea.Batch(textarea.Blink, m.loadCommits())
		}
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleAskInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateSmartMenu
		m.askInput.Blur()
		m.err = nil
		return m, nil
	case "ctrl+enter", "ctrl+g":
		query := m.askInput.Value()
		if query == "" {
			m.err = fmt.Errorf("please enter a query")
			return m, nil
		}
		m.askQuery = query
		m.state = stateGenerating
		m.status = "Claude is thinking..."
		return m, m.generateFromQuery()
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+t":
		m.allowThread = !m.allowThread
		return m, nil
	}
	var cmd tea.Cmd
	m.askInput, cmd = m.askInput.Update(msg)
	return m, cmd
}

func (m Model) handleCommitBrowserKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const maxVisible = 5

	switch msg.String() {
	case "esc":
		if m.commitSearchActive {
			m.commitSearchActive = false
			m.commitSearch = ""
			m.filterCommits()
			m.commitCursor = 0
			m.commitScrollOffset = 0
			return m, nil
		}
		if m.commitPromptActive {
			m.commitPromptActive = false
			m.commitPromptInput.Blur()
			return m, nil
		}
		m.state = stateSmartMenu
		m.commits = nil
		m.selectedCommits = nil
		m.err = nil
		return m, nil
	case "q":
		if m.commitSearchActive || m.commitPromptActive {
			// Pass to input
		} else {
			m.state = stateSmartMenu
			m.commits = nil
			m.selectedCommits = nil
			m.err = nil
			return m, nil
		}
	case "/":
		if m.commitPromptActive {
			var cmd tea.Cmd
			m.commitPromptInput, cmd = m.commitPromptInput.Update(msg)
			return m, cmd
		}
		if m.commitSearchActive {
			m.commitSearch += "/"
			m.filterCommits()
			m.commitCursor = 0
			m.commitScrollOffset = 0
			return m, nil
		}
		if m.commitSearch != "" {
			m.commitSearch = ""
			m.filterCommits()
			m.commitCursor = 0
			m.commitScrollOffset = 0
		} else {
			m.commitSearchActive = true
		}
		return m, nil
	case "tab":
		if m.commitSearchActive {
			m.commitSearchActive = false
			return m, nil
		}
		m.commitPromptActive = !m.commitPromptActive
		if m.commitPromptActive {
			m.commitPromptInput.Focus()
			return m, textarea.Blink
		} else {
			m.commitPromptInput.Blur()
			return m, nil
		}
	case "up", "k":
		if m.commitPromptActive {
			var cmd tea.Cmd
			m.commitPromptInput, cmd = m.commitPromptInput.Update(msg)
			return m, cmd
		}
		if m.commitSearchActive {
			break
		}
		if m.commitCursor > 0 {
			m.commitCursor--
			if m.commitCursor < m.commitScrollOffset {
				m.commitScrollOffset = m.commitCursor
			}
		}
	case "down", "j":
		if m.commitPromptActive {
			var cmd tea.Cmd
			m.commitPromptInput, cmd = m.commitPromptInput.Update(msg)
			return m, cmd
		}
		if m.commitSearchActive {
			break
		}
		if m.commitCursor < len(m.filteredCommits)-1 {
			m.commitCursor++
			if m.commitCursor >= m.commitScrollOffset+maxVisible {
				m.commitScrollOffset = m.commitCursor - maxVisible + 1
			}
		}
	case "a":
		if m.commitSearchActive {
			m.commitSearch += "a"
			m.filterCommits()
			m.commitCursor = 0
			m.commitScrollOffset = 0
			return m, nil
		}
		if m.commitPromptActive {
			var cmd tea.Cmd
			m.commitPromptInput, cmd = m.commitPromptInput.Update(msg)
			return m, cmd
		}
		// Toggle select all filtered commits
		if len(m.selectedCommits) == len(m.filteredCommits) {
			m.selectedCommits = nil
		} else {
			m.selectedCommits = make([]int, len(m.filteredCommits))
			copy(m.selectedCommits, m.filteredCommits)
		}
		return m, nil
	case " ":
		if m.commitSearchActive {
			m.commitSearch += " "
			m.filterCommits()
			return m, nil
		}
		if m.commitPromptActive {
			var cmd tea.Cmd
			m.commitPromptInput, cmd = m.commitPromptInput.Update(msg)
			return m, cmd
		}
		if m.commitCursor < len(m.filteredCommits) {
			realIdx := m.filteredCommits[m.commitCursor]
			found := false
			for i, s := range m.selectedCommits {
				if s == realIdx {
					m.selectedCommits = append(m.selectedCommits[:i], m.selectedCommits[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				m.selectedCommits = append(m.selectedCommits, realIdx)
			}
		}
	case "enter", "ctrl+g":
		if m.commitSearchActive {
			m.commitSearchActive = false
			return m, nil
		}
		if m.commitPromptActive {
			m.commitPromptActive = false
			m.commitPromptInput.Blur()
		}
		if len(m.selectedCommits) == 0 && m.commitCursor < len(m.filteredCommits) {
			m.selectedCommits = []int{m.filteredCommits[m.commitCursor]}
		}
		m.state = stateGenerating
		m.status = "Generating suggestion..."
		return m, m.generateSuggestion()
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+t":
		m.allowThread = !m.allowThread
		return m, nil
	case "backspace":
		if m.commitSearchActive && len(m.commitSearch) > 0 {
			m.commitSearch = m.commitSearch[:len(m.commitSearch)-1]
			m.filterCommits()
			m.commitCursor = 0
			m.commitScrollOffset = 0
			return m, nil
		}
		if m.commitPromptActive {
			var cmd tea.Cmd
			m.commitPromptInput, cmd = m.commitPromptInput.Update(msg)
			return m, cmd
		}
	default:
		if m.commitSearchActive {
			if len(msg.String()) == 1 {
				m.commitSearch += msg.String()
				m.filterCommits()
				m.commitCursor = 0
				m.commitScrollOffset = 0
			}
			return m, nil
		}
		if m.commitPromptActive {
			var cmd tea.Cmd
			m.commitPromptInput, cmd = m.commitPromptInput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

// handleComposeKeys handles keys for both stateCompose and stateSmartCompose
func (m Model) handleComposeKeys(msg tea.KeyMsg, isSmartPost bool) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if isSmartPost {
			m.state = stateSmartMenu
		} else {
			m.state = stateHome
			m.thread = []threadItem{{text: "", mediaIDs: nil, media: nil}}
			m.currentPost = 0
		}
		m.textarea.Blur()
		m.err = nil
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

	case "ctrl+r":
		if isSmartPost {
			m.state = stateGenerating
			m.status = "Regenerating..."
			return m, m.generateSuggestion()
		}

	case "ctrl+o":
		if len(m.thread[m.currentPost].media) >= 4 {
			m.err = fmt.Errorf("maximum 4 images per post")
			return m, nil
		}
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

	case "ctrl+d":
		if len(m.thread) > 1 {
			m.thread = append(m.thread[:m.currentPost], m.thread[m.currentPost+1:]...)
			if m.currentPost >= len(m.thread) {
				m.currentPost = len(m.thread) - 1
			}
			m.textarea.SetValue(m.thread[m.currentPost].text)
		}
		return m, nil

	case "ctrl+x":
		if !isSmartPost && len(m.thread[m.currentPost].media) > 0 {
			m.thread[m.currentPost].media = m.thread[m.currentPost].media[:len(m.thread[m.currentPost].media)-1]
			m.thread[m.currentPost].mediaIDs = m.thread[m.currentPost].mediaIDs[:len(m.thread[m.currentPost].mediaIDs)-1]
		}
		return m, nil

	case "ctrl+left", "ctrl+b":
		if isSmartPost && m.currentPost > 0 {
			m.thread[m.currentPost].text = m.textarea.Value()
			m.currentPost--
			m.textarea.SetValue(m.thread[m.currentPost].text)
		}
		return m, nil

	case "ctrl+right", "ctrl+f":
		if isSmartPost && m.currentPost < len(m.thread)-1 {
			m.thread[m.currentPost].text = m.textarea.Value()
			m.currentPost++
			m.textarea.SetValue(m.thread[m.currentPost].text)
		}
		return m, nil

	case "ctrl+up", "ctrl+k":
		if !isSmartPost && m.currentPost > 0 {
			m.thread[m.currentPost].text = m.textarea.Value()
			m.currentPost--
			m.textarea.SetValue(m.thread[m.currentPost].text)
		}
		return m, nil

	case "ctrl+down", "ctrl+j":
		if !isSmartPost && m.currentPost < len(m.thread)-1 {
			m.thread[m.currentPost].text = m.textarea.Value()
			m.currentPost++
			m.textarea.SetValue(m.thread[m.currentPost].text)
		}
		return m, nil

	case "ctrl+c":
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m Model) handleMediaInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		path := strings.TrimSpace(m.pathInput.Value())
		if path == "" {
			if m.isSmartPost {
				m.state = stateSmartCompose
			} else {
				m.state = stateCompose
			}
			m.textarea.Focus()
			return m, textarea.Blink
		}
		// Clean up paths from drag-and-drop (escaped spaces, quotes)
		path = strings.ReplaceAll(path, "\\ ", " ")
		path = strings.Trim(path, "'\"")
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
	return m, cmd
}

func (m Model) handlePostedKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	return m, nil
}

// Helper methods

func (m *Model) filterCommits() {
	if m.commitSearch == "" {
		m.filteredCommits = make([]int, len(m.commits))
		for i := range m.commits {
			m.filteredCommits[i] = i
		}
		return
	}

	search := strings.ToLower(m.commitSearch)
	m.filteredCommits = nil
	for i, commit := range m.commits {
		if strings.Contains(strings.ToLower(commit.Subject), search) {
			m.filteredCommits = append(m.filteredCommits, i)
		}
	}
}

func (m Model) hasContent() bool {
	for _, item := range m.thread {
		if strings.TrimSpace(item.text) != "" || len(item.mediaIDs) > 0 {
			return true
		}
	}
	return strings.TrimSpace(m.textarea.Value()) != ""
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
