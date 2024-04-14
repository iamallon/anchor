package storage

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

const (
	stdUser = "git"
)

var (
	ErrInvalidURL = errors.New("not a valid URL")
)

type gitStorage struct {
	path string
	auth transport.AuthMethod
}

func newGitStorage(path string) (*gitStorage, error) {
	auth, err := ssh.NewSSHAgentAuth(stdUser)
	if err != nil {
		return nil, err
	}

	return &gitStorage{
		path: path,
		auth: auth,
	}, nil
}

func (storage *gitStorage) Init(args ...string) error {
	if len(args) == 0 {
		return ErrInvalidURL
	}

	_, err := git.PlainClone(storage.path, false, &git.CloneOptions{
		URL:  args[0],
		Auth: storage.auth,
	})

	return err
}

func (storage *gitStorage) Update() error {
	repo, err := git.PlainOpen(storage.path)
	if err != nil {
		return err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = tree.Pull(&git.PullOptions{
		RemoteName: git.DefaultRemoteName,
		Auth:       storage.auth,
	})

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}

func (storage *gitStorage) Store(msg string) error {
	repo, err := git.PlainOpen(storage.path)
	if err != nil {
		return err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = tree.Add(".")
	if err != nil {
		return err
	}

	_, err = tree.Commit(msg, &git.CommitOptions{})
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: git.DefaultRemoteName,
		Auth:       storage.auth,
	})

	return err
}

func (storage *gitStorage) Diff() (string, error) {
	repo, err := git.PlainOpen(storage.path)
	if err != nil {
		return "", err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	status, err := tree.Status()
	if err != nil {
		return "", err
	}

	return status.String(), nil
}
