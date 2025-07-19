package command

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/loghinalexandru/anchor/internal/command/util/label"
	"github.com/loghinalexandru/anchor/internal/config"
	"github.com/loghinalexandru/anchor/internal/model"
	"github.com/loghinalexandru/anchor/internal/output"
	"github.com/loghinalexandru/anchor/internal/output/bubbletea"
	"github.com/loghinalexandru/anchor/internal/output/bubbletea/style"
	"github.com/peterbourgon/ff/v4"
)

const (
	viewName      = "view"
	viewUsage     = "anchor view [LABEL]"
	viewShortHelp = "view and edit existing bookmarks"
	viewLongHelp  = `  This command will open up the interactive TUI that can view/edit each individual bookmark stored with [LABEL].
  Prompts for confirmation for any change on exit.

EXAMPLES
  # View bookmarks under label "programming"
  anchor view programming

  # View bookmarks with sub-label "go" under parent label "programming"
  anchor view programming go
`
)

const (
	msgApplyChanges = "You are about to apply changes from previous operation. Proceed?"
)

type viewCmd struct{}

func (v *viewCmd) manifest(parent *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("view").SetParent(parent)

	return &ff.Command{
		Name:      viewName,
		Usage:     viewUsage,
		ShortHelp: viewShortHelp,
		LongHelp:  viewLongHelp,
		Flags:     flags,
		Exec: func(ctx context.Context, args []string) error {
			return v.handle(ctx.(appContext), args)
		},
	}
}

func (v *viewCmd) handle(ctx appContext, args []string) (err error) {
	fh, err := label.OpenFuzzy(config.DataDirPath(), args, os.O_RDWR)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, fh.Close())
	}()

	var bookmarks []list.Item
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		bk, err := model.BookmarkLine(scanner.Text())
		if err != nil {
			return err
		}

		bookmarks = append(bookmarks, bk)
	}

	runner := tea.NewProgram(bubbletea.NewView(bookmarks, filepath.Base(fh.Name())), tea.WithContext(ctx))
	state, err := runner.Run()
	if err != nil {
		return err
	}

	view := state.(*bubbletea.View)
	if !view.Dirty() {
		return nil
	}

	confirmer := output.Confirmer{
		MaxRetries: 3,
		Renderer:   style.Prompt,
	}
	if !confirmer.Confirm(msgApplyChanges, os.Stdin, os.Stdout) {
		return nil
	}

	err = fh.Truncate(0)
	if err != nil {
		return err
	}

	_, err = fh.Seek(0, 0)
	if err != nil {
		return err
	}

	for _, b := range view.Bookmarks() {
		_, err := fh.WriteString(b.String())
		if err != nil {
			return err
		}
	}

	for _, a := range view.Actions() {
		switch a.Operation {
		case bubbletea.Delete:
			// Explicitly ignore if there is a remove error.
			_ = os.Remove(config.ArchiveFilePath(a.Target))
		}
	}

	return nil
}
