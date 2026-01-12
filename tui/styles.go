package tui

import "github.com/charmbracelet/lipgloss"

// ANSI color constants - these use the terminal's color palette
// so they automatically match whatever theme the user has
const (
	ansiBlack        = lipgloss.Color("0")
	ansiRed          = lipgloss.Color("1")
	ansiGreen        = lipgloss.Color("2")
	ansiYellow       = lipgloss.Color("3")
	ansiBlue         = lipgloss.Color("4")
	ansiMagenta      = lipgloss.Color("5")
	ansiCyan         = lipgloss.Color("6")
	ansiWhite        = lipgloss.Color("7")
	ansiBrightBlack  = lipgloss.Color("8")
	ansiBrightRed    = lipgloss.Color("9")
	ansiBrightGreen  = lipgloss.Color("10")
	ansiBrightYellow = lipgloss.Color("11")
	ansiBrightBlue   = lipgloss.Color("12")
	ansiBrightMagenta = lipgloss.Color("13")
	ansiBrightCyan   = lipgloss.Color("14")
	ansiBrightWhite  = lipgloss.Color("15")
)

// Style variables
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
	initStyles()
}

func initStyles() {
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ansiBrightRed)

	taglineStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack).
		Italic(true)

	subtitleStyle = lipgloss.NewStyle().
		Foreground(ansiBrightWhite).
		Bold(true)

	menuItemStyle = lipgloss.NewStyle().
		Foreground(ansiWhite)

	menuDescStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack).
		PaddingLeft(4)

	selectedStyle = lipgloss.NewStyle().
		Foreground(ansiBrightYellow).
		Bold(true)

	selectedDescStyle = lipgloss.NewStyle().
		Foreground(ansiMagenta).
		PaddingLeft(4)

	bulletStyle = lipgloss.NewStyle().
		Foreground(ansiBrightRed).
		Bold(true)

	dimBulletStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack)

	disabledStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack)

	disabledTagStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack).
		Padding(0, 1)

	disabledDescStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack).
		PaddingLeft(4)

	helpBarStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack).
		Border(lipgloss.Border{Top: "─"}).
		BorderForeground(ansiBrightBlack).
		PaddingTop(1).
		MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(ansiBlack).
		Background(ansiCyan).
		Padding(0, 1).
		Bold(true)

	helpTextStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack)

	statusStyle = lipgloss.NewStyle().
		Foreground(ansiBrightGreen).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(ansiBrightRed).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(ansiBrightYellow)

	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ansiBrightBlack).
		Padding(0, 1)

	activeBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ansiBrightCyan).
		Padding(0, 1)

	urlStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlue).
		Underline(true)

	mediaTagStyle = lipgloss.NewStyle().
		Foreground(ansiBlack).
		Background(ansiGreen).
		Padding(0, 1)

	threadNumStyle = lipgloss.NewStyle().
		Foreground(ansiBlack).
		Background(ansiBlue).
		Padding(0, 1).
		Bold(true)

	dimStyle = lipgloss.NewStyle().
		Foreground(ansiBrightBlack)

	inputLabelStyle = lipgloss.NewStyle().
		Foreground(ansiWhite)

	commitHashStyle = lipgloss.NewStyle().
		Foreground(ansiBrightGreen)

	commitTimeStyle = lipgloss.NewStyle().
		Foreground(ansiBrightMagenta)

	aiTagStyle = lipgloss.NewStyle().
		Foreground(ansiBlack).
		Background(ansiMagenta).
		Padding(0, 1).
		Bold(true)
}
