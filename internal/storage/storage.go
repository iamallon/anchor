package storage

import (
	"strings"

	"github.com/loghinalexandru/anchor/internal/config"
)

type Kind int

const (
	Local Kind = iota
	Git
)

type Storer interface {
	Init(args ...string) error
	Store(msg string) error
}

func New(k Kind) Storer {
	switch k {
	case Git:
		storer, err := newGitStorage(config.DataDirPath())
		if err != nil {
			panic(err)
		}

		return storer
	default:
		return newLocalStorage(config.DataDirPath())
	}
}

func Parse(s string) Kind {
	switch strings.ToLower(s) {
	case "git":
		return Git
	default:
		return Local
	}
}
