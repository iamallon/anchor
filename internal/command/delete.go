package command

import (
	"context"

	"github.com/loghinalexandru/anchor/internal/command/util/label"
	"github.com/loghinalexandru/anchor/internal/config"
	"github.com/loghinalexandru/anchor/internal/output"
	"github.com/peterbourgon/ff/v4"
)

const (
	deleteName      = "delete"
	deleteUsage     = "anchor delete [FLAGS]"
	deleteShortHelp = "remove all bookmarks under specified labels"
	deleteLongHelp  = `  Performs a bulk delete on all the bookmarks under the specified labels.
  Prompts for confirmation before deleting.`
)

const (
	msgDeleteLabel = "You are about to delete the label and associated items. Proceed?"
)

type deleteCmd struct{}

func (del *deleteCmd) manifest(parent *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("delete").SetParent(parent)

	return &ff.Command{
		Name:      deleteName,
		Usage:     deleteUsage,
		ShortHelp: deleteShortHelp,
		LongHelp:  deleteLongHelp,
		Flags:     flags,
		Exec: func(ctx context.Context, args []string) error {
			return del.handle(ctx.(appContext), args)
		},
	}
}

func (del *deleteCmd) handle(_ appContext, args []string) (err error) {
	ok := output.Confirm(msgDeleteLabel)
	if !ok {
		return nil
	}

	return label.Remove(config.DataDirPath(), args)
}
