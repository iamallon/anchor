package command

import (
	"context"
	"fmt"
	"os"

	"github.com/loghinalexandru/anchor/internal/config"
	"github.com/loghinalexandru/anchor/internal/output/treeprint"
	"github.com/peterbourgon/ff/v4"
)

const (
	treeName      = "tree"
	treeUsage     = "anchor tree"
	treeShortHelp = "list available labels in a tree structure"
	treeLongHelp  = `  Print to stdout a tree like structure to see exactly the current label hierarchy.
  The values on the left of each label represents the number of distinct bookmarks it holds.`
)

type treeCmd struct{}

func (tree *treeCmd) manifest(parent *ff.FlagSet) *ff.Command {
	return &ff.Command{
		Name:      treeName,
		Usage:     treeUsage,
		ShortHelp: treeShortHelp,
		LongHelp:  treeLongHelp,
		Flags:     ff.NewFlagSet("tree").SetParent(parent),
		Exec: func(ctx context.Context, args []string) error {
			return tree.handle(ctx.(appContext), args)
		},
	}
}

func (*treeCmd) handle(ctx appContext, _ []string) error {
	dd := os.DirFS(config.DataDirPath())
	fmt.Print(treeprint.Generate(dd))

	return nil
}
