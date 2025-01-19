package bubbletea

import (
	"fmt"
	"os/exec"
	"runtime"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/loghinalexandru/anchor/internal/model"
	"github.com/loghinalexandru/anchor/internal/output/bubbletea/style"
)

const (
	msgStatus = "Deleted %q"
)

var (
	quitKey    = key.NewBinding(key.WithKeys("esc", "q"))
	confirmKey = key.NewBinding(key.WithKeys("enter", " "))
	delKey     = key.NewBinding(key.WithKeys("d", "delete"))
	renameKey  = key.NewBinding(key.WithKeys("r"))
	startKey   = key.NewBinding(key.WithKeys("home"))
	endKey     = key.NewBinding(key.WithKeys("end"))
)

type View struct {
	input     textinput.Model
	bookmarks list.Model
	dirty     bool
}

func NewView(bookmarks []list.Item, title string) *View {
	del := list.NewDefaultDelegate()
	style.ApplyToDelegate(&del)

	viewList := list.New(bookmarks, del, 0, 0)
	style.ApplyToList(title, &viewList)

	input := textinput.New()
	input.KeyMap.LineStart = startKey
	input.KeyMap.LineEnd = endKey

	return &View{
		input:     input,
		bookmarks: viewList,
	}
}

func (v *View) Bookmarks() []*model.Bookmark {
	res := make([]*model.Bookmark, len(v.bookmarks.Items()))

	for i, bk := range v.bookmarks.Items() {
		res[i] = bk.(*model.Bookmark)
	}

	return res
}

func (v *View) Dirty() bool {
	return v.dirty
}

func (v *View) Init() tea.Cmd {
	return nil
}

func (v *View) View() string {
	if v.input.Focused() {
		v.bookmarks.SetShowPagination(false)
		v.bookmarks.SetShowHelp(false)
		return style.Default().Render(v.bookmarks.View() + "\n" + v.input.View())
	}

	v.bookmarks.SetShowPagination(true)
	v.bookmarks.SetShowHelp(true)
	return style.Default().Render(v.bookmarks.View())
}

func (v *View) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var inputCmd tea.Cmd
	var viewCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		x, y := style.Default().GetFrameSize()
		width := msg.Width - x
		height := msg.Height - y
		v.input.Width = width - x
		v.bookmarks.SetSize(width, height)
		v.bookmarks, inputCmd = v.bookmarks.Update(msg)
		v.input, viewCmd = v.input.Update(msg)
	case tea.KeyMsg:
		// Bypass "msg" input pipeline if setting filter value.
		if v.bookmarks.SettingFilter() {
			v.bookmarks, viewCmd = v.bookmarks.Update(msg)
			break
		}

		if v.input.Focused() {
			v.input, inputCmd = v.handleInput(msg)
			break
		}

		v.bookmarks, viewCmd = v.handleList(msg)
	default:
		v.input, inputCmd = v.input.Update(msg)
		v.bookmarks, viewCmd = v.bookmarks.Update(msg)
	}

	return v, tea.Batch(inputCmd, viewCmd)
}

func (v *View) handleInput(msg tea.KeyMsg) (textinput.Model, tea.Cmd) {
	if key.Matches(msg, quitKey) || key.Matches(msg, confirmKey) {
		item := v.bookmarks.SelectedItem().(*model.Bookmark)
		if v.input.Value() != item.Title() {
			item.Update(v.input.Value())
			v.dirty = true
		}

		v.input.Reset()
		v.input.Blur()
		return v.input, tea.ClearScreen
	}

	return v.input.Update(msg)
}

func (v *View) handleList(msg tea.KeyMsg) (list.Model, tea.Cmd) {
	item, ok := v.bookmarks.SelectedItem().(*model.Bookmark)
	if !ok {
		return v.bookmarks.Update(msg)
	}

	switch {
	case key.Matches(msg, confirmKey):
		_ = open(item.URL())
	case key.Matches(msg, delKey):
		var cmd tea.Cmd
		items := slices.DeleteFunc(v.bookmarks.Items(), func(item list.Item) bool {
			return item == v.bookmarks.SelectedItem()
		})

		if len(items) != len(v.bookmarks.Items()) {
			v.dirty = true
			cmd = tea.Batch(v.bookmarks.SetItems(items), v.bookmarks.NewStatusMessage(fmt.Sprintf(msgStatus, item.Title())))
		}

		return v.bookmarks, cmd
	case key.Matches(msg, renameKey):
		v.input.SetValue(item.Title())
		v.input.Focus()
		return v.bookmarks, textinput.Blink
	}

	return v.bookmarks.Update(msg)
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
