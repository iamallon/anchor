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
	msgDeleteLabel = "You are about to delete the label and associated bookmarks. Proceed?"
)

type deleteCmd struct {
	labels []string
}

func (del *deleteCmd) manifest(parent *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("delete").SetParent(parent)
	flags.StringSetVar(&del.labels, 'l', "label", "add label in order of appearance")

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

func (del *deleteCmd) handle(_ appContext, _ []string) (err error) {
	ok := output.Confirm(msgDeleteLabel)
	if !ok {
		return nil
	}

	return label.Remove(config.DataDirPath(), del.labels)
}
