package command

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/loghinalexandru/anchor/internal/config"
	"github.com/loghinalexandru/anchor/internal/output"
	"github.com/loghinalexandru/anchor/internal/storage"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffyaml"
)

const (
	rootName        = "anchor"
	rootUsage       = "anchor <SUBCOMMAND>"
	msgUpdateFailed = "Failed pulling latest changes. Continue operation?"
)

type Updater interface {
	Update() error
}

// appContext is a context.Context wrapper for
// type safety and to avoid key-value pairs.
type appContext struct {
	context.Context
	kind     storage.Kind
	storer   storage.Storer
	syncMode string
	client   *http.Client
	scraper  *colly.Collector
	template *template.Template
}

type rootCmd struct {
	cmd *ff.Command
}

func newRoot() *rootCmd {
	root := &rootCmd{}

	rootFlags := ff.NewFlagSet("anchor")

	root.cmd = &ff.Command{
		Name:  rootName,
		Usage: rootUsage,
		Flags: rootFlags,
	}

	root.cmd.Subcommands = []*ff.Command{
		(&initCmd{}).manifest(rootFlags),
		(&viewCmd{}).manifest(rootFlags),
		(&addCmd{}).manifest(rootFlags),
		(&deleteCmd{}).manifest(rootFlags),
		(&treeCmd{}).manifest(rootFlags),
		(&syncCmd{}).manifest(rootFlags),
		(&importCmd{}).manifest(rootFlags),
		(&versionCmd{}).manifest(rootFlags),
	}

	return root
}

func (root *rootCmd) handle(ctx context.Context, args []string) (err error) {
	err = root.cmd.Parse(args)
	if err != nil {
		return err
	}

	fh, err := os.Open(config.SettingsFilePath())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	defer func() {
		if fh != nil {
			err = errors.Join(err, fh.Close())
		}
	}()

	// Configure global scraper.
	// This might be a bad idea because the instance is shared
	// and API has side effects like registering a callback.
	scraper := colly.NewCollector()
	scraper.IgnoreRobotsTxt = true

	// Configure global HTML template.
	style := []string{
		"max-width: 40em",
		"margin-right: 10%",
		"margin-left: 10%",
		"margin-top:7%",
		"margin-bottom: 7%",
	}
	tmpl, _ := template.New("root").Parse(fmt.Sprintf(`<div style="%s;">{{.}}</div>`, strings.Join(style, ";")))

	// Initialize appContext with sensible defaults.
	appCtx := appContext{
		Context:  ctx,
		kind:     storage.Local,
		syncMode: "always",
		client:   &http.Client{Timeout: config.StdHttpTimeout},
		scraper:  scraper,
		template: tmpl,
	}

	// Config file might not exist, ignore errors if so.
	_ = ffyaml.Parse(fh, func(key, value string) error {
		switch key {
		case config.StdSyncModeKey:
			appCtx.syncMode = value
		case config.StdStorageKey:
			appCtx.kind = storage.Parse(value)
		}

		return nil
	})

	// Initialize storer after config was read to not miss
	// any custom values e.g. path.
	appCtx.storer = storage.New(appCtx.kind)

	// Add appropriate middleware for each subcommand
	for _, c := range root.cmd.Subcommands {
		switch c.Name {
		// Skip updateMiddleware for commands that
		// do not need to fetch from remote.
		case initName:
			c.Exec = contextMiddleware(c.Exec)
		case versionName:
			c.Exec = contextMiddleware(c.Exec)
		default:
			c.Exec = contextMiddleware(updaterMiddleware(c.Exec, appCtx))
		}
	}

	return root.cmd.Run(appCtx)
}

type handlerFunc func(ctx context.Context, args []string) error

func updaterMiddleware(next handlerFunc, appCtx appContext) handlerFunc {
	updater, ok := appCtx.storer.(Updater)
	if !ok || appCtx.syncMode != "always" {
		return next
	}

	return func(ctx context.Context, args []string) error {
		err := updater.Update()
		if err != nil {
			if ok := output.Confirm(msgUpdateFailed); !ok {
				return err
			}
		}

		return next(ctx, args)
	}
}

func contextMiddleware(next handlerFunc) handlerFunc {
	return func(ctx context.Context, args []string) error {
		res := make(chan error, 1)

		go func(res chan<- error) {
			res <- next(ctx, args)
			close(res)
		}(res)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-res:
			return err
		}
	}
}
