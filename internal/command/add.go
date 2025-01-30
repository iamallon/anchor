package command

import (
	"context"
	"errors"
	"html/template"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/loghinalexandru/anchor/internal/command/util/label"
	"github.com/loghinalexandru/anchor/internal/command/util/parser"
	"github.com/loghinalexandru/anchor/internal/config"
	"github.com/loghinalexandru/anchor/internal/model"
	"github.com/peterbourgon/ff/v4"
)

const (
	addName      = "add"
	addUsage     = "anchor add [FLAGS]"
	addShortHelp = "append a bookmark entry with set labels"
	addLongHelp  = `  Append a bookmark to a file on the backing storage determined by the 
  flatten hierarchy of the provided labels. Order of the flags matter when storing the entry.

  If no label is provided via the -l flag, all the entries will be added
  to the default "root" label.

  By default it tries to fetch the "title" content from the provided URL. If it fails
  to do so, it will store the entry with same title as the URL. You can provide a specific
  title with the flag -t and it overwrites the behaviour mentioned above.

  You can also provider a comment via the flag -c for the bookmark instead of the default URL that is specified. 

EXAMPLES
  # Append to default label
  anchor add "https://www.youtube.com/"

  # Append to a label "programming" with a sub-label "go"
  anchor add -l programming -l go "https://gobyexample.com/"
  anchor add -l go "https://go.dev/ref/spec" -c "GO: Language Spec"
`
)

type addCmd struct {
	labels  []string
	title   string
	comment string
	expr    string
}

func (add *addCmd) manifest(parent *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("add").SetParent(parent)
	flags.StringSetVar(&add.labels, 'l', "label", "add labels in order of appearance")
	flags.StringVar(&add.title, 't', "title", "", "add custom title")
	flags.StringVar(&add.comment, 'c', "comment", "", "add bookmark comment")
	// Let users specify a CSS selector since every page is different.
	flags.StringVar(&add.expr, 'e', "expr", "", "add a CSS style selector to scrape the bookmark")

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

	if add.expr != "" {
		_ = scrapeAndStore(add.expr, b, ctx)
	}

	return nil
}

// Do proper error handling
func scrapeAndStore(expr string, b *model.Bookmark, ctx appContext) error {
	ctx.scraper.OnHTML("img", func(el *colly.HTMLElement) {
		val, _ := el.DOM.Attr("src")
		src, _ := url.Parse(val)
		if src.IsAbs() {
			return
		}
		absolute, _ := url.JoinPath(b.URL(), src.String())
		el.DOM.SetAttr("src", absolute)
	})

	ctx.scraper.OnHTML(expr, func(el *colly.HTMLElement) {
		file := path.Join(config.DataDirPath(), strings.ReplaceAll(b.Title()+".html", " ", "_"))
		fh, _ := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, config.StdFileMode)
		content, _ := el.DOM.Html()
		_ = ctx.template.Execute(fh, template.HTML(content))
		_ = fh.Close()
	})

	return ctx.scraper.Visit(b.URL())
}
