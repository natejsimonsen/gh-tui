package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name        string
	Fg          lipgloss.Color
	FgSubtle    lipgloss.Color
	FgMuted     lipgloss.Color
	BgHighlight lipgloss.Color
	Green       lipgloss.Color
	Red         lipgloss.Color
	Yellow      lipgloss.Color
	Purple      lipgloss.Color
	Blue        lipgloss.Color
	Cyan        lipgloss.Color
	Orange      lipgloss.Color
	Border      lipgloss.Color
}

var themes = []Theme{
	{
		Name: "Catppuccin Mocha", Fg: "#cdd6f4", FgSubtle: "#a6adc8", FgMuted: "#6c7086",
		BgHighlight: "#45475a", Green: "#a6e3a1", Red: "#f38ba8", Yellow: "#f9e2af",
		Purple: "#cba6f7", Blue: "#89b4fa", Cyan: "#94e2d5", Orange: "#fab387", Border: "#585b70",
	},
	{
		Name: "Catppuccin Latte", Fg: "#4c4f69", FgSubtle: "#6c6f85", FgMuted: "#9ca0b0",
		BgHighlight: "#bcc0cc", Green: "#40a02b", Red: "#d20f39", Yellow: "#df8e1d",
		Purple: "#8839ef", Blue: "#1e66f5", Cyan: "#179299", Orange: "#fe640b", Border: "#acb0be",
	},
	{
		Name: "Dracula", Fg: "#f8f8f2", FgSubtle: "#6272a4", FgMuted: "#44475a",
		BgHighlight: "#44475a", Green: "#50fa7b", Red: "#ff5555", Yellow: "#f1fa8c",
		Purple: "#bd93f9", Blue: "#8be9fd", Cyan: "#8be9fd", Orange: "#ffb86c", Border: "#6272a4",
	},
	{
		Name: "Tokyo Night", Fg: "#c0caf5", FgSubtle: "#565f89", FgMuted: "#3b4261",
		BgHighlight: "#292e42", Green: "#9ece6a", Red: "#f7768e", Yellow: "#e0af68",
		Purple: "#bb9af7", Blue: "#7aa2f7", Cyan: "#7dcfff", Orange: "#ff9e64", Border: "#3b4261",
	},
	{
		Name: "Nord", Fg: "#d8dee9", FgSubtle: "#4c566a", FgMuted: "#3b4252",
		BgHighlight: "#3b4252", Green: "#a3be8c", Red: "#bf616a", Yellow: "#ebcb8b",
		Purple: "#b48ead", Blue: "#81a1c1", Cyan: "#88c0d0", Orange: "#d08770", Border: "#4c566a",
	},
	{
		Name: "Gruvbox Dark", Fg: "#ebdbb2", FgSubtle: "#928374", FgMuted: "#665c54",
		BgHighlight: "#3c3836", Green: "#b8bb26", Red: "#fb4934", Yellow: "#fabd2f",
		Purple: "#d3869b", Blue: "#83a598", Cyan: "#8ec07c", Orange: "#fe8019", Border: "#665c54",
	},
	{
		Name: "Solarized Dark", Fg: "#839496", FgSubtle: "#586e75", FgMuted: "#073642",
		BgHighlight: "#073642", Green: "#859900", Red: "#dc322f", Yellow: "#b58900",
		Purple: "#6c71c4", Blue: "#268bd2", Cyan: "#2aa198", Orange: "#cb4b16", Border: "#586e75",
	},
	{
		Name: "Solarized Light", Fg: "#657b83", FgSubtle: "#93a1a1", FgMuted: "#eee8d5",
		BgHighlight: "#eee8d5", Green: "#859900", Red: "#dc322f", Yellow: "#b58900",
		Purple: "#6c71c4", Blue: "#268bd2", Cyan: "#2aa198", Orange: "#cb4b16", Border: "#93a1a1",
	},
	{
		Name: "One Dark", Fg: "#abb2bf", FgSubtle: "#5c6370", FgMuted: "#3e4451",
		BgHighlight: "#3e4451", Green: "#98c379", Red: "#e06c75", Yellow: "#e5c07b",
		Purple: "#c678dd", Blue: "#61afef", Cyan: "#56b6c2", Orange: "#d19a66", Border: "#5c6370",
	},
	{
		Name: "Rosé Pine", Fg: "#e0def4", FgSubtle: "#908caa", FgMuted: "#6e6a86",
		BgHighlight: "#26233a", Green: "#31748f", Red: "#eb6f92", Yellow: "#f6c177",
		Purple: "#c4a7e7", Blue: "#9ccfd8", Cyan: "#9ccfd8", Orange: "#f6c177", Border: "#6e6a86",
	},
}

func applyTheme(t Theme) {
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Purple)
	openStyle = lipgloss.NewStyle().Foreground(t.Green)
	mergedStyle = lipgloss.NewStyle().Foreground(t.Purple)
	closedStyle = lipgloss.NewStyle().Foreground(t.Red)
	draftStyle = lipgloss.NewStyle().Foreground(t.FgMuted)
	ciPassStyle = lipgloss.NewStyle().Foreground(t.Green)
	ciFailStyle = lipgloss.NewStyle().Foreground(t.Red)
	ciPendStyle = lipgloss.NewStyle().Foreground(t.Yellow)
	headerStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	helpStyle = lipgloss.NewStyle().Foreground(t.FgSubtle)
	statusStyle = lipgloss.NewStyle().Foreground(t.FgSubtle).Padding(0, 1)
	metaStyle = lipgloss.NewStyle().Foreground(t.FgSubtle)
	authorStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Purple)
	timeStyle = lipgloss.NewStyle().Foreground(t.FgMuted)
	dividerStyle = lipgloss.NewStyle().Foreground(t.Border)
	diffAddStyle = lipgloss.NewStyle().Foreground(t.Green)
	diffDelStyle = lipgloss.NewStyle().Foreground(t.Red)
	diffHunkStyle = lipgloss.NewStyle().Foreground(t.Cyan)
	diffFileStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Blue)
}
