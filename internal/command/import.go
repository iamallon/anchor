package command

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/loghinalexandru/anchor/internal/command/util/parser"
	"github.com/loghinalexandru/anchor/internal/config"
	"github.com/peterbourgon/ff/v4"
	"github.com/virtualtam/netscape-go/v2"
)

const (
	importName      = "import"
	importUsage     = "anchor import [PATH]"
	importShortHelp = "import bookmarks from a browser exported file"
	importLongHelp  = `  Imports all the bookmarks from a "NETSCAPE-Bookmark-file-1" file format setting up the appropriate labels
  based on the folder structure that was previously in the browser.

  On import, it formats all the invalid folder names because they will be reused as labels inside anchor.
  Valid label names contain only lower case alphanumeric characters and hyphen.
`
)

var (
	ErrInvalidImportFile = errors.New("invalid import file")
)

type importCmd struct{}

func (imp *importCmd) manifest(parent *ff.FlagSet) *ff.Command {
	return &ff.Command{
		Name:      importName,
		Usage:     importUsage,
		ShortHelp: importShortHelp,
		LongHelp:  importLongHelp,
		Flags:     ff.NewFlagSet("import").SetParent(parent),
		Exec: func(ctx context.Context, args []string) error {
			return imp.handle(ctx.(appContext), args)
		},
	}
}

func (*importCmd) handle(_ appContext, args []string) error {
	if len(args) == 0 {
		return ErrInvalidImportFile
	}

	fh, err := os.Open(args[0])
	if err != nil {
		return err
	}

	content, err := io.ReadAll(fh)
	if err != nil {
		return err
	}

	doc, _ := netscape.Unmarshal(content)
	err = parser.TraverseNode(config.DataDirPath(), nil, doc.Root)

	if err != nil {
		return err
	}

	return nil
}
