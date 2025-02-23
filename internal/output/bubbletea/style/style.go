package style

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/lipgloss"
)

type RenderFunc func(in string) string

var (
	stdStyle       = lipgloss.NewStyle().Margin(2, 2, 2, 2)
	stdPromptStyle = lipgloss.NewStyle().Margin(0, 0, 0, 2)
)

func Nop(in string) string {
	return in
}

func Prompt(in string) string {
	return stdPromptStyle.Render(in)
}

func Default() lipgloss.Style {
	return stdStyle
}

func ApplyToDelegate(del *list.DefaultDelegate) {
	del.Styles.SelectedTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.NoColor{})
	del.Styles.SelectedDesc = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.NoColor{})
}

func ApplyToList(title string, list *list.Model) {
	list.Title = title
	list.InfiniteScrolling = true
	list.Paginator.Type = paginator.Arabic
	list.Paginator.ArabicFormat = "%d/%d \u2693"
	list.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.NoColor{})
	list.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.NoColor{})

	list.KeyMap.PrevPage.SetKeys("left", "h", "pgup")
	list.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	// This is duplicated in the commands as well. Need to move
	// in a central config place maybe.
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter", "space"), key.WithHelp("enter", "open")),
			key.NewBinding(key.WithKeys("delete", "d"), key.WithHelp("d/del", "delete")),
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
		}
	}

	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter", "space"), key.WithHelp("enter/space", "open in browser")),
			key.NewBinding(key.WithKeys("delete", "d"), key.WithHelp("d/del", "remove bookmark")),
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename bookmark")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "view archived page")),
		}
	}
}
