package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/gocolly/colly/v2"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func archive(e *colly.HTMLElement) {
	_, _ = fmt.Println(e.DOM.Text())
}

func Execute(args []string) error {
	root := newRoot()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := root.handle(ctx, args)
	if errors.Is(err, ff.ErrHelp) || errors.Is(err, ff.ErrNoExec) {
		_, err = fmt.Fprint(os.Stdout, ffhelp.Command(root.cmd))
		return err
	}

	return err
}
