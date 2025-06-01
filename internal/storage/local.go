package storage

import (
	"errors"
	"os"
)

type localStorage struct {
	path string
}

func newLocalStorage(path string) *localStorage {
	return &localStorage{
		path: path,
	}
}

func (l *localStorage) Init(_ ...string) error {
	if _, err := os.Stat(l.path); os.IsNotExist(err) {
		err = os.MkdirAll(l.path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*localStorage) Diff() (string, error) {
	return "", errors.New("running on local storage type, command has no effect")
}

func (*localStorage) Store(_ string) error {
	return nil
}
