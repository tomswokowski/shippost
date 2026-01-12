package tui

import "github.com/charmbracelet/lipgloss"

// Adaptive colors - automatically pick light/dark variant based on terminal background
var (
	colorTitle        = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#FF6B6B"}
	colorTagline      = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#6B7280"}
	colorSubtitle     = lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#E2E8F0"}
	colorText         = lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#E2E8F0"}
	colorDim          = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#64748B"}
	colorDimmer       = lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#475569"}
	colorSelected     = lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FFE66D"}
	colorSelectedDesc = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	colorBullet       = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#FF6B6B"}
	colorBorder       = lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#334155"}
	colorBorderActive = lipgloss.AdaptiveColor{Light: "#0D9488", Dark: "#4ECDC4"}
	colorHelpKey      = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#0C4A6E"}
	colorHelpKeyBg    = lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#0EA5E9"}
	colorSuccess      = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"}
	colorError        = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"}
	colorWarning      = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#F59E0B"}
	colorURL          = lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"}
	colorMediaTag     = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#10B981"}
	colorMediaTagBg   = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#064E3B"}
	colorThreadNum    = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#6366F1"}
	colorThreadNumBg  = lipgloss.AdaptiveColor{Light: "#4F46E5", Dark: "#312E81"}
	colorCommitHash   = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#95E6CB"}
	colorCommitTime   = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	colorAITag        = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#F472B6"}
	colorAITagBg      = lipgloss.AdaptiveColor{Light: "#DB2777", Dark: "#831843"}
	colorInputLabel   = lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#94A3B8"}
	colorDisabledBg   = lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#1E293B"}
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
		Foreground(colorTitle)

	taglineStyle = lipgloss.NewStyle().
		Foreground(colorTagline).
		Italic(true)

	subtitleStyle = lipgloss.NewStyle().
		Foreground(colorSubtitle).
		Bold(true)

	menuItemStyle = lipgloss.NewStyle().
		Foreground(colorText)

	menuDescStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		PaddingLeft(4)

	selectedStyle = lipgloss.NewStyle().
		Foreground(colorSelected).
		Bold(true)

	selectedDescStyle = lipgloss.NewStyle().
		Foreground(colorSelectedDesc).
		PaddingLeft(4)

	bulletStyle = lipgloss.NewStyle().
		Foreground(colorBullet).
		Bold(true)

	dimBulletStyle = lipgloss.NewStyle().
		Foreground(colorDimmer)

	disabledStyle = lipgloss.NewStyle().
		Foreground(colorDimmer)

	disabledTagStyle = lipgloss.NewStyle().
		Foreground(colorDimmer).
		Background(colorDisabledBg).
		Padding(0, 1)

	disabledDescStyle = lipgloss.NewStyle().
		Foreground(colorDimmer).
		PaddingLeft(4)

	helpBarStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		Border(lipgloss.Border{Top: "─"}).
		BorderForeground(colorBorder).
		PaddingTop(1).
		MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(colorHelpKey).
		Background(colorHelpKeyBg).
		Padding(0, 1).
		Bold(true)

	helpTextStyle = lipgloss.NewStyle().
		Foreground(colorDim)

	statusStyle = lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(colorError).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(colorWarning)

	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1)

	activeBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorderActive).
		Padding(0, 1)

	urlStyle = lipgloss.NewStyle().
		Foreground(colorURL).
		Underline(true)

	mediaTagStyle = lipgloss.NewStyle().
		Foreground(colorMediaTag).
		Background(colorMediaTagBg).
		Padding(0, 1)

	threadNumStyle = lipgloss.NewStyle().
		Foreground(colorThreadNum).
		Background(colorThreadNumBg).
		Padding(0, 1).
		Bold(true)

	dimStyle = lipgloss.NewStyle().
		Foreground(colorDim)

	inputLabelStyle = lipgloss.NewStyle().
		Foreground(colorInputLabel)

	commitHashStyle = lipgloss.NewStyle().
		Foreground(colorCommitHash)

	commitTimeStyle = lipgloss.NewStyle().
		Foreground(colorCommitTime)

	aiTagStyle = lipgloss.NewStyle().
		Foreground(colorAITag).
		Background(colorAITagBg).
		Padding(0, 1).
		Bold(true)
}
