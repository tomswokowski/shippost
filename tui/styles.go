package tui

import "github.com/charmbracelet/lipgloss"

// Style variables - initialized in init() based on terminal theme
var (
	titleStyle        lipgloss.Style
	taglineStyle      lipgloss.Style
	subtitleStyle     lipgloss.Style
	menuItemStyle     lipgloss.Style
	menuDescStyle     lipgloss.Style
	selectedStyle     lipgloss.Style
	selectedDescStyle lipgloss.Style
	bulletStyle       lipgloss.Style
	dimBulletStyle    lipgloss.Style
	disabledStyle     lipgloss.Style
	disabledTagStyle  lipgloss.Style
	disabledDescStyle lipgloss.Style
	helpBarStyle      lipgloss.Style
	helpKeyStyle      lipgloss.Style
	helpTextStyle     lipgloss.Style
	statusStyle       lipgloss.Style
	errorStyle        lipgloss.Style
	warningStyle      lipgloss.Style
	boxStyle          lipgloss.Style
	activeBoxStyle    lipgloss.Style
	urlStyle          lipgloss.Style
	mediaTagStyle     lipgloss.Style
	threadNumStyle    lipgloss.Style
	dimStyle          lipgloss.Style
	inputLabelStyle   lipgloss.Style
	commitHashStyle   lipgloss.Style
	commitTimeStyle   lipgloss.Style
	aiTagStyle        lipgloss.Style
)

func init() {
	isDark := lipgloss.HasDarkBackground()
	initStyles(isDark)
}

func initStyles(isDark bool) {
	if isDark {
		initDarkStyles()
	} else {
		initLightStyles()
	}
}

func initDarkStyles() {
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
}

func initLightStyles() {
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#DC2626"))

	taglineStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	subtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1F2937")).
		Bold(true)

	menuItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1F2937"))

	menuDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		PaddingLeft(4)

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B45309")).
		Bold(true)

	selectedDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		PaddingLeft(4)

	bulletStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DC2626")).
		Bold(true)

	dimBulletStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))

	disabledStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))

	disabledTagStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Background(lipgloss.Color("#F3F4F6")).
		Padding(0, 1)

	disabledDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		PaddingLeft(4)

	helpBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Border(lipgloss.Border{Top: "─"}).
		BorderForeground(lipgloss.Color("#D1D5DB")).
		PaddingTop(1).
		MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2563EB")).
		Padding(0, 1).
		Bold(true)

	helpTextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DC2626")).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D97706"))

	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#D1D5DB")).
		Padding(0, 1)

	activeBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0D9488")).
		Padding(0, 1)

	urlStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2563EB")).
		Underline(true)

	mediaTagStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#059669")).
		Padding(0, 1)

	threadNumStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4F46E5")).
		Padding(0, 1).
		Bold(true)

	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	inputLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4B5563"))

	commitHashStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669"))

	commitTimeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED"))

	aiTagStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#DB2777")).
		Padding(0, 1).
		Bold(true)
}
