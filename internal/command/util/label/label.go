package label

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/loghinalexandru/anchor/internal/config"
	"github.com/loghinalexandru/anchor/internal/model"
	"github.com/sahilm/fuzzy"
)

var (
	ErrInvalidLabel = errors.New("invalid label passed")
	ErrMissingLabel = errors.New("missing file for label(s) passed")
)

var notLabelRegexp = regexp.MustCompile(`([^a-z0-9-]|^$)`)

// Open validates and opens the file constructed from the labels.
// If the file does not exist or some labels are invalid, returns ErrMissingLabel
// and ErrInvalidLabel respectively.
// If os.O_CREATE flag is passed and the file does not exist,
// it is created with mode perm config.StdFileMode.
func Open(rootDir string, labels []string, flag int) (*os.File, error) {
	err := validate(labels)
	if err != nil {
		return nil, err
	}

	fh, err := os.OpenFile(filepath.Join(rootDir, filename(labels)), flag, config.StdFileMode)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrMissingLabel
	}

	return fh, err
}

// OpenFuzzy performs the same action as Open but picks the
// file based on a ranked fuzzy match on given labels.
func OpenFuzzy(rootDir string, labels []string, flag int) (*os.File, error) {
	err := validate(labels)
	if err != nil {
		return nil, err
	}

	fname := match(rootDir, labels)
	fh, err := os.OpenFile(filepath.Join(rootDir, fname), flag, config.StdFileMode)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, ErrMissingLabel
	}

	return fh, err
}

// Remove validates and removes the file constructed from the labels.
// If the file does not exist, has no effect.
func Remove(rootDir string, labels []string) error {
	err := validate(labels)
	if err != nil {
		return err
	}

	fh, err := os.Open(filepath.Join(rootDir, filename(labels)))
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		bk, _ := model.BookmarkLine(scanner.Text())
		_ = os.Remove(config.ArchiveFilePath(bk.Id()))
	}

	_ = fh.Close()

	err = os.Remove(filepath.Join(rootDir, filename(labels)))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return nil
}

// Format formats and strips out any invalid characters from provided labels.
// Invalid character is anything that [^a-z0-9-] does match.
func Format(labels []string) []string {
	result := make([]string, len(labels))
	for i, l := range labels {
		result[i] = notLabelRegexp.ReplaceAllString(strings.ToLower(l), "")
	}

	return result
}

func match(rootDir string, labels []string) string {
	var matchData []string

	dd, err := os.ReadDir(rootDir)
	if err != nil {
		return filename(labels)
	}

	for _, d := range dd {
		if !d.IsDir() {
			matchData = append(matchData, d.Name())
		}
	}

	target := filename(labels)
	matches := fuzzy.Find(target, matchData)
	if len(matches) <= 0 {
		return target
	}

	return matches[0].Str
}

func filename(labels []string) string {
	if len(labels) == 0 {
		return config.StdLabel
	}

	return strings.Join(labels, config.StdLabelSeparator)
}

func validate(labels []string) error {
	var result error
	for _, l := range labels {
		if notLabelRegexp.MatchString(l) {
			result = errors.Join(result, fmt.Errorf("%q: %w", l, ErrInvalidLabel))
		}
	}

	return result
}
