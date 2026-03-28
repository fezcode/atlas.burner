package tui

import "github.com/charmbracelet/lipgloss"

var (
	// ── Colors ──────────────────────────────────────────────────
	colorPrimary   = lipgloss.Color("#FF6F00") // warm amber
	colorSecondary = lipgloss.Color("#FFB74D") // light amber
	colorAccent    = lipgloss.Color("#FF3D00") // fiery red-orange
	colorSuccess   = lipgloss.Color("#66BB6A") // green
	colorDanger    = lipgloss.Color("#EF5350") // red
	colorMuted     = lipgloss.Color("#9E9E9E") // gray
	colorDim       = lipgloss.Color("#616161") // dark gray
	colorSurface   = lipgloss.Color("#1E1E1E") // dark bg
	colorBorder    = lipgloss.Color("#424242") // subtle border
	colorWhite     = lipgloss.Color("#FAFAFA")

	// ── Layout ──────────────────────────────────────────────────
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// ── Header / Banner ─────────────────────────────────────────
	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Italic(true)

	// ── Step indicator ──────────────────────────────────────────
	stepIndicatorStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				MarginBottom(1)

	activeStepStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	inactiveStepStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	completedStepStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Strikethrough(true)

	// ── List items ──────────────────────────────────────────────
	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			PaddingLeft(4)

	itemDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			PaddingLeft(4)

	// ── Progress bar ────────────────────────────────────────────
	progressFullStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	// ── Panels / Boxes ──────────────────────────────────────────
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2).
			MarginTop(1)

	confirmPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(colorDanger).
				Padding(1, 2).
				MarginTop(1)

	successPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorSuccess).
				Padding(1, 2).
				MarginTop(1)

	// ── Inline ──────────────────────────────────────────────────
	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	dangerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorDanger)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSuccess)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)
)

// banner returns the ASCII art header
func banner() string {
	art := `
   ▄▀█ ▀█▀ █   ▄▀█ █▀   ██▄ █ █ █▀█ █▄ █ █▀▀ █▀█
   █▀█  █  █▄▄ █▀█ ▄█   █▄█ █▄█ █▀▄ █ ▀█ ██▄ █▀▄`
	return bannerStyle.Render(art)
}
