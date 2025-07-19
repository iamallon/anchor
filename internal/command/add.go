package command

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"os"

	readability "github.com/go-shiori/go-readability"
	"github.com/loghinalexandru/anchor/internal/command/util/label"
	"github.com/loghinalexandru/anchor/internal/command/util/parser"
	"github.com/loghinalexandru/anchor/internal/config"
	"github.com/loghinalexandru/anchor/internal/model"
	"github.com/peterbourgon/ff/v4"
)

const (
	addName      = "add"
	addUsage     = "anchor add [FLAGS] <URL>"
	addShortHelp = "append a bookmark entry with set labels"
	addLongHelp  = `  Append a bookmark to a file on the backing storage determined by the 
  flatten hierarchy of the provided labels. Order of the flags matter when storing the entry.

  If no label is provided via the -l flag, all the entries will be added
  to the default "root" label.

  By default it tries to fetch the "title" content from the provided URL. If it fails
  to do so, it will store the entry with same title as the URL. You can provide a specific
  title with the flag -t and it overwrites the behaviour mentioned above.

  You can also provide a comment via the flag -c for the bookmark instead of the default URL that is specified. 

  If you wish to store locally the target page you can specify the -e flag with a CSS style expression
  and it will fetch and store a simplified version of the page locally.

EXAMPLES
  # Append to default label
  anchor add "https://www.youtube.com/"

  # Append to a label "programming" with a sub-label "go"
  anchor add -l programming -l go "https://gobyexample.com/"
  anchor add -l go -c "GO: Language Spec" "https://go.dev/ref/spec"
  anchor add -l go -a "https://go.dev/ref/spec"
`
)

type addCmd struct {
	labels  []string
	title   string
	comment string
	archive bool
}

func (add *addCmd) manifest(parent *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("add").SetParent(parent)
	flags.StringSetVar(&add.labels, 'l', "label", "add labels in order of appearance")
	flags.StringVar(&add.title, 't', "title", "", "add custom title")
	flags.StringVar(&add.comment, 'c', "comment", "", "add bookmark comment")
	flags.BoolVar(&add.archive, 'a', "archive", "store a local copy")

	return &ff.Command{
		Name:      addName,
		Usage:     addUsage,
		ShortHelp: addShortHelp,
		LongHelp:  addLongHelp,
		Flags:     flags,
		Exec: func(ctx context.Context, args []string) error {
			return add.handle(ctx.(appContext), args)
		},
	}
}

func (add *addCmd) handle(ctx appContext, args []string) error {
	target := parser.First(args)

	b, err := model.NewBookmark(
		target,
		model.WithTitle(add.title),
		model.WithClient(ctx.client),
		model.WithComment(add.comment))
	if err != nil {
		return err
	}

	file, err := label.Open(config.DataDirPath(), add.labels, os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}

	err = b.Write(file)
	err = errors.Join(err, file.Close())
	if err != nil {
		return err
	}

	if add.archive {
		filePath := config.ArchiveFilePath(b.Id())
		fh, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, config.StdFileMode)
		if err != nil {
			return fmt.Errorf("could not open file, make sure you run the `init` command first")
		}

		content, err := readability.FromURL(b.URL(), config.StdHttpTimeout)
		if err != nil {
			return err
		}

		err = ctx.template.Execute(fh, template.HTML(content.Content))
		errors.Join(err, fh.Close())
		if err != nil {
			return err
		}
	}

	return nil
}
