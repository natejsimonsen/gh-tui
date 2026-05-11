package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name        string
	Fg          lipgloss.Color
	FgSubtle    lipgloss.Color
	FgMuted     lipgloss.Color
	BgHighlight lipgloss.Color
	SelectedBg  lipgloss.Color
	Green       lipgloss.Color
	Red         lipgloss.Color
	Yellow      lipgloss.Color
	Purple      lipgloss.Color
	Blue        lipgloss.Color
	Cyan        lipgloss.Color
	Orange      lipgloss.Color
	Border      lipgloss.Color
	DiffAddBg   lipgloss.Color
	DiffDelBg   lipgloss.Color
	DiffAddFg   lipgloss.Color
	DiffDelFg   lipgloss.Color
}

var themes = []Theme{
	{
		Name: "Catppuccin Mocha", Fg: "#cdd6f4", FgSubtle: "#a6adc8", FgMuted: "#6c7086",
		BgHighlight: "#45475a", SelectedBg: "#313450", Green: "#a6e3a1", Red: "#f38ba8", Yellow: "#f9e2af",
		Purple: "#cba6f7", Blue: "#89b4fa", Cyan: "#94e2d5", Orange: "#fab387", Border: "#585b70",
		DiffAddBg: "#1e3a2a", DiffDelBg: "#3a1e2a", DiffAddFg: "#a6e3a1", DiffDelFg: "#f38ba8",
	},
	{
		Name: "Catppuccin Latte", Fg: "#4c4f69", FgSubtle: "#6c6f85", FgMuted: "#9ca0b0",
		BgHighlight: "#bcc0cc", SelectedBg: "#c9cde0", Green: "#40a02b", Red: "#d20f39", Yellow: "#df8e1d",
		Purple: "#8839ef", Blue: "#1e66f5", Cyan: "#179299", Orange: "#fe640b", Border: "#acb0be",
		DiffAddBg: "#d5ecd5", DiffDelBg: "#f5d5d5", DiffAddFg: "#2a6e2a", DiffDelFg: "#9e1a1a",
	},
	{
		Name: "Dracula", Fg: "#f8f8f2", FgSubtle: "#6272a4", FgMuted: "#44475a",
		BgHighlight: "#44475a", SelectedBg: "#3d3f58", Green: "#50fa7b", Red: "#ff5555", Yellow: "#f1fa8c",
		Purple: "#bd93f9", Blue: "#8be9fd", Cyan: "#8be9fd", Orange: "#ffb86c", Border: "#6272a4",
		DiffAddBg: "#1a3a1a", DiffDelBg: "#3a1a1a", DiffAddFg: "#50fa7b", DiffDelFg: "#ff5555",
	},
	{
		Name: "Tokyo Night", Fg: "#c0caf5", FgSubtle: "#565f89", FgMuted: "#3b4261",
		BgHighlight: "#292e42", SelectedBg: "#283457", Green: "#9ece6a", Red: "#f7768e", Yellow: "#e0af68",
		Purple: "#bb9af7", Blue: "#7aa2f7", Cyan: "#7dcfff", Orange: "#ff9e64", Border: "#3b4261",
		DiffAddBg: "#1a2e1a", DiffDelBg: "#2e1a1e", DiffAddFg: "#9ece6a", DiffDelFg: "#f7768e",
	},
	{
		Name: "Nord", Fg: "#d8dee9", FgSubtle: "#4c566a", FgMuted: "#3b4252",
		BgHighlight: "#3b4252", SelectedBg: "#3b4963", Green: "#a3be8c", Red: "#bf616a", Yellow: "#ebcb8b",
		Purple: "#b48ead", Blue: "#81a1c1", Cyan: "#88c0d0", Orange: "#d08770", Border: "#4c566a",
		DiffAddBg: "#2a3a2a", DiffDelBg: "#3a2a2a", DiffAddFg: "#a3be8c", DiffDelFg: "#bf616a",
	},
	{
		Name: "Gruvbox Dark", Fg: "#ebdbb2", FgSubtle: "#928374", FgMuted: "#665c54",
		BgHighlight: "#3c3836", SelectedBg: "#3d4220", Green: "#b8bb26", Red: "#fb4934", Yellow: "#fabd2f",
		Purple: "#d3869b", Blue: "#83a598", Cyan: "#8ec07c", Orange: "#fe8019", Border: "#665c54",
		DiffAddBg: "#2a3020", DiffDelBg: "#3a2020", DiffAddFg: "#b8bb26", DiffDelFg: "#fb4934",
	},
	{
		Name: "Solarized Dark", Fg: "#839496", FgSubtle: "#586e75", FgMuted: "#073642",
		BgHighlight: "#073642", SelectedBg: "#0a3a50", Green: "#859900", Red: "#dc322f", Yellow: "#b58900",
		Purple: "#6c71c4", Blue: "#268bd2", Cyan: "#2aa198", Orange: "#cb4b16", Border: "#586e75",
		DiffAddBg: "#0a2a1a", DiffDelBg: "#2a0a0a", DiffAddFg: "#859900", DiffDelFg: "#dc322f",
	},
	{
		Name: "Solarized Light", Fg: "#657b83", FgSubtle: "#93a1a1", FgMuted: "#eee8d5",
		BgHighlight: "#eee8d5", SelectedBg: "#ddd8c5", Green: "#859900", Red: "#dc322f", Yellow: "#b58900",
		Purple: "#6c71c4", Blue: "#268bd2", Cyan: "#2aa198", Orange: "#cb4b16", Border: "#93a1a1",
		DiffAddBg: "#d5e8c5", DiffDelBg: "#f0d0d0", DiffAddFg: "#556b00", DiffDelFg: "#b01a1a",
	},
	{
		Name: "One Dark", Fg: "#abb2bf", FgSubtle: "#5c6370", FgMuted: "#3e4451",
		BgHighlight: "#3e4451", SelectedBg: "#2c3545", Green: "#98c379", Red: "#e06c75", Yellow: "#e5c07b",
		Purple: "#c678dd", Blue: "#61afef", Cyan: "#56b6c2", Orange: "#d19a66", Border: "#5c6370",
		DiffAddBg: "#1e2e1e", DiffDelBg: "#2e1e1e", DiffAddFg: "#98c379", DiffDelFg: "#e06c75",
	},
	{
		Name: "Rosé Pine", Fg: "#e0def4", FgSubtle: "#908caa", FgMuted: "#6e6a86",
		BgHighlight: "#26233a", SelectedBg: "#2a2844", Green: "#31748f", Red: "#eb6f92", Yellow: "#f6c177",
		Purple: "#c4a7e7", Blue: "#9ccfd8", Cyan: "#9ccfd8", Orange: "#f6c177", Border: "#6e6a86",
		DiffAddBg: "#1a2a30", DiffDelBg: "#30202a", DiffAddFg: "#31748f", DiffDelFg: "#eb6f92",
	},
}

var activeTheme Theme

func applyTheme(t Theme) {
	activeTheme = t
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
	diffAddStyle = lipgloss.NewStyle().Foreground(t.DiffAddFg).Background(t.DiffAddBg)
	diffDelStyle = lipgloss.NewStyle().Foreground(t.DiffDelFg).Background(t.DiffDelBg)
	diffHunkStyle = lipgloss.NewStyle().Foreground(t.Cyan).Faint(true)
	diffFileStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Blue)
	diffFileBannerStyle = lipgloss.NewStyle().Bold(true).Foreground(t.Fg).Background(t.BgHighlight).Padding(0, 1)
	diffGutterStyle = lipgloss.NewStyle().Foreground(t.FgMuted)
	diffCtxStyle = lipgloss.NewStyle().Foreground(t.FgSubtle)
}
