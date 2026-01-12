package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	minTerminalHeight = 20
	minTerminalWidth  = 60
)

// View renders the current state of the TUI
func (m Model) View() string {
	// Check for minimum terminal size
	if m.height > 0 && m.height < minTerminalHeight {
		return m.viewTooSmall()
	}
	if m.width > 0 && m.width < minTerminalWidth {
		return m.viewTooSmall()
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("shippost"))
	b.WriteString("  ")
	b.WriteString(taglineStyle.Render("Share your work with the world"))
	b.WriteString("\n\n")

	switch m.state {
	case stateHome:
		m.viewHome(&b)
	case stateSmartMenu:
		m.viewSmartMenu(&b)
	case stateAskInput:
		m.viewAskInput(&b)
	case stateCommitBrowser:
		m.viewCommitBrowser(&b)
	case stateGenerating:
		m.viewGenerating(&b)
	case stateSmartCompose:
		m.viewCompose(&b, true)
	case stateCompose, statePosting:
		m.viewCompose(&b, false)
	case stateMediaInput:
		m.viewMediaInput(&b)
	case statePosted:
		m.viewPosted(&b)
	}

	return b.String()
}

func (m Model) viewTooSmall() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("shippost"))
	b.WriteString("\n\n")
	b.WriteString(warningStyle.Render("Terminal too small"))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("Please resize to at least %d×%d", minTerminalWidth, minTerminalHeight)))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("Current size: %d×%d", m.width, m.height)))
	return b.String()
}

func (m Model) viewHome(b *strings.Builder) {
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
}

func (m Model) viewSmartMenu(b *strings.Builder) {
	b.WriteString(subtitleStyle.Render("Smart Post"))
	b.WriteString("\n\n")

	smartMenuItems := []struct {
		title string
		desc  string
	}{
		{"Browse Commits", "Pick specific commits to post about"},
		{"Ask", "Describe what you want to post about"},
	}

	for i, item := range smartMenuItems {
		if i == m.smartMenuCursor {
			b.WriteString(bulletStyle.Render("▸ "))
			b.WriteString(selectedStyle.Render(item.title))
			b.WriteString("\n")
			b.WriteString(selectedDescStyle.Render(item.desc))
		} else {
			b.WriteString(dimBulletStyle.Render("  "))
			b.WriteString(menuItemStyle.Render(item.title))
			b.WriteString("\n")
			b.WriteString(menuDescStyle.Render(item.desc))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(m.renderHelpBar([]helpItem{
		{"↑↓", "navigate"},
		{"enter", "select"},
		{"esc", "back"},
	}))
}

func (m Model) viewAskInput(b *strings.Builder) {
	b.WriteString(subtitleStyle.Render("Smart Post"))
	b.WriteString("  ")
	b.WriteString(aiTagStyle.Render(" Ask "))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("What would you like to post about?"))
	b.WriteString("\n\n")

	b.WriteString(activeBoxStyle.Render(m.askInput.View()))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n\n")
	}

	m.renderThreadToggle(b)

	b.WriteString(dimStyle.Render("Examples:"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  • What did I accomplish today?"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  • What good practices am I using?"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  • Summarize my recent refactoring work"))
	b.WriteString("\n\n")

	b.WriteString(m.renderHelpBar([]helpItem{
		{"ctrl+enter", "generate"},
		{"ctrl+t", "single/thread"},
		{"esc", "back"},
	}))
}

func (m Model) viewCommitBrowser(b *strings.Builder) {
	maxVisible := m.commitListMaxVisible()

	b.WriteString(subtitleStyle.Render("Smart Post"))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("Select commits to post about"))
	b.WriteString("\n\n")

	// Search bar
	if m.commitSearchActive {
		b.WriteString(dimStyle.Render("/"))
		b.WriteString(selectedStyle.Render(m.commitSearch))
		b.WriteString(selectedStyle.Render("▌"))
		b.WriteString("\n\n")
	} else if m.commitSearch != "" {
		b.WriteString(dimStyle.Render("/"))
		b.WriteString(menuItemStyle.Render(m.commitSearch))
		b.WriteString("  ")
		b.WriteString(dimStyle.Render(fmt.Sprintf("(%d matches)", len(m.filteredCommits))))
		b.WriteString("\n\n")
	}

	if m.status != "" {
		b.WriteString(statusStyle.Render("● " + m.status))
		b.WriteString("\n")
	} else if m.err != nil {
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n")
	} else if len(m.commits) == 0 {
		b.WriteString(dimStyle.Render("No commits found in this repository"))
	} else if len(m.filteredCommits) == 0 {
		b.WriteString(dimStyle.Render("No matching commits"))
	} else {
		// Show scroll indicator if there are more commits above
		if m.commitScrollOffset > 0 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  ↑ %d more above\n", m.commitScrollOffset)))
		}

		// Show visible window of commits
		end := m.commitScrollOffset + maxVisible
		if end > len(m.filteredCommits) {
			end = len(m.filteredCommits)
		}

		for i := m.commitScrollOffset; i < end; i++ {
			realIdx := m.filteredCommits[i]
			commit := m.commits[realIdx]

			isSelected := false
			for _, s := range m.selectedCommits {
				if s == realIdx {
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

		// Show scroll indicator if there are more commits below
		remaining := len(m.filteredCommits) - end
		if remaining > 0 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  ↓ %d more below\n", remaining)))
		}

		if len(m.selectedCommits) > 0 {
			b.WriteString("\n")
			b.WriteString(dimStyle.Render(fmt.Sprintf("%d commit(s) selected", len(m.selectedCommits))))
		}

		// Prompt input
		b.WriteString("\n\n")
		b.WriteString(inputLabelStyle.Render("Prompt "))
		b.WriteString(dimStyle.Render("(optional)"))
		b.WriteString("\n")
		if m.commitPromptActive {
			b.WriteString(activeBoxStyle.Render(m.commitPromptInput.View()))
		} else {
			b.WriteString(boxStyle.Render(m.commitPromptInput.View()))
		}
	}

	// Thread mode toggle (always visible)
	b.WriteString("\n\n")
	m.renderThreadToggle(b)

	b.WriteString("\n")
	searchHelp := "search"
	if !m.commitSearchActive && m.commitSearch != "" {
		searchHelp = "clear"
	}
	b.WriteString(m.renderHelpBar([]helpItem{
		{"↑↓", "navigate"},
		{"space", "select"},
		{"/", searchHelp},
		{"tab", "prompt"},
		{"ctrl+t", "single/thread"},
		{"enter", "generate"},
		{"esc", "back"},
	}))
}

func (m Model) viewGenerating(b *strings.Builder) {
	b.WriteString(subtitleStyle.Render("Smart Post"))
	b.WriteString("\n\n")
	b.WriteString(statusStyle.Render("● " + m.status))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Claude is writing your post..."))
}

// viewCompose renders both Quick Post and Smart Post compose screens
func (m Model) viewCompose(b *strings.Builder, isSmartPost bool) {
	if isSmartPost {
		b.WriteString(subtitleStyle.Render("Smart Post"))
		b.WriteString("  ")
		b.WriteString(aiTagStyle.Render(" AI "))
	} else {
		b.WriteString(subtitleStyle.Render("Quick Post"))
	}

	if len(m.thread) > 1 {
		b.WriteString("  ")
		b.WriteString(threadNumStyle.Render(fmt.Sprintf(" %s %d/%d ", m.threadLabel(isSmartPost), m.currentPost+1, len(m.thread))))
	}
	b.WriteString("\n")

	// Thread indicator dots
	if len(m.thread) > 1 {
		b.WriteString("\n")
		for i := range m.thread {
			if i == m.currentPost {
				b.WriteString(selectedStyle.Render("●"))
			} else {
				b.WriteString(dimStyle.Render("○"))
			}
			if i < len(m.thread)-1 {
				b.WriteString(dimStyle.Render("─"))
			}
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Textarea
	if m.state == statePosting {
		b.WriteString(boxStyle.Render(m.textarea.View()))
	} else {
		b.WriteString(activeBoxStyle.Render(m.textarea.View()))
	}
	b.WriteString("\n")

	// Character count
	m.renderCharCount(b)

	// Media tags
	if len(m.thread[m.currentPost].media) > 0 {
		b.WriteString("  ")
		for _, path := range m.thread[m.currentPost].media {
			b.WriteString(mediaTagStyle.Render(" 📎 " + filepath.Base(path) + " "))
			b.WriteString(" ")
		}
	}

	// Error
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
	}

	// Status for posting state
	if m.state == statePosting {
		b.WriteString("\n")
		b.WriteString(statusStyle.Render("● " + m.status))
	}

	// Thread preview for Quick Post with multiple posts
	if !isSmartPost && len(m.thread) > 1 {
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
	b.WriteString(m.renderHelpBar(m.composeHelpItems(isSmartPost)))
}

func (m Model) viewMediaInput(b *strings.Builder) {
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
}

func (m Model) viewPosted(b *strings.Builder) {
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

// Helper methods for views

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

func (m Model) renderThreadToggle(b *strings.Builder) {
	if m.allowThread {
		b.WriteString(selectedStyle.Render("● "))
		b.WriteString(menuItemStyle.Render("Allow threads"))
		b.WriteString(dimStyle.Render(" (default)"))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("○ "))
		b.WriteString(dimStyle.Render("Single post only"))
	} else {
		b.WriteString(dimStyle.Render("○ "))
		b.WriteString(dimStyle.Render("Allow threads"))
		b.WriteString(dimStyle.Render(" (default)"))
		b.WriteString("\n")
		b.WriteString(selectedStyle.Render("● "))
		b.WriteString(menuItemStyle.Render("Single post only"))
	}
	b.WriteString("\n\n")
}

func (m Model) renderCharCount(b *strings.Builder) {
	charCount := utf8.RuneCountInString(m.textarea.Value())
	countStyle := helpTextStyle
	if charCount > 260 {
		countStyle = warningStyle
	}
	if charCount > 280 {
		countStyle = errorStyle
	}
	b.WriteString(countStyle.Render(fmt.Sprintf("%d", charCount)))
	b.WriteString(helpTextStyle.Render("/280"))
}

func (m Model) threadLabel(isSmartPost bool) string {
	if isSmartPost {
		return "THREAD"
	}
	return ""
}

func (m Model) composeHelpItems(isSmartPost bool) []helpItem {
	items := []helpItem{
		{"ctrl+s", "send"},
	}
	if isSmartPost {
		items = append(items, helpItem{"ctrl+r", "regen"})
	}
	items = append(items, helpItem{"ctrl+o", "attach"})
	items = append(items, helpItem{"ctrl+n", "add"})

	if len(m.thread[m.currentPost].media) > 0 && !isSmartPost {
		items = append(items, helpItem{"ctrl+x", "remove media"})
	}
	if len(m.thread) > 1 {
		items = append(items, helpItem{"ctrl+d", "delete"})
		if isSmartPost {
			items = append(items, helpItem{"ctrl+←→", "nav"})
		} else {
			items = append(items, helpItem{"ctrl+↑↓", "nav"})
		}
	}
	items = append(items, helpItem{"esc", "back"})
	return items
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
