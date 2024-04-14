package treeprint_test

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/loghinalexandru/anchor/internal/output/treeprint"
)

func TestGenerateWithDir(t *testing.T) {
	t.Parallel()

	want := "anchor\n"
	got := treeprint.Generate(fstest.MapFS{
		"git": {
			Mode: fs.ModeDir,
		},
	})

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("output not matching; (-got, +want):\n %s", diff)
	}
}

func TestGenerateEmpty(t *testing.T) {
	t.Parallel()

	want := "anchor\n"
	got := treeprint.Generate(fstest.MapFS{})

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("output not matching; (-got, +want):\n %s", diff)
	}
}

func TestGenerateSimple(t *testing.T) {
	t.Parallel()

	want, err := os.ReadFile("testdata/simple.golden")
	if err != nil {
		t.Fatalf("unexpected error; got %q", err)
	}

	fsys := fstest.MapFS{
		"root":            {},
		"prometheus.algo": {},
		"prometheus":      {},
		"prometheus.go":   {},
	}

	got := treeprint.Generate(fsys)

	if diff := cmp.Diff(got, string(want)); diff != "" {
		t.Log(got)
		t.Errorf("output not matching; (-got, +want):\n %s", diff)
	}
}

func TestGenerateComplex(t *testing.T) {
	t.Parallel()

	want, err := os.ReadFile("testdata/complex.golden")
	if err != nil {
		t.Fatalf("unexpected error; got %q", err)
	}

	fsys := fstest.MapFS{
		"a":         {},
		"a.b.c":     {},
		"a.b.d":     {},
		"a.c":       {},
		"a.d.b":     {},
		"a.d.g.f.y": {},
		"c.b.a":     {},
		"x.y":       {},
	}

	got := treeprint.Generate(fsys)

	if diff := cmp.Diff(got, string(want)); diff != "" {
		t.Log(got)
		t.Errorf("output not matching; (-got, +want):\n %s", diff)
	}
}

func TestGenerateWithMetadata(t *testing.T) {
	t.Parallel()

	want, err := os.ReadFile("testdata/metadata.golden")
	if err != nil {
		t.Fatalf("unexpected error; got %q", err)
	}

	fsys := fstest.MapFS{
		"root": {
			Data: []byte("first-line\nsecond-line\n"),
		},
		"prometheus.go": {
			Data: []byte("first-line\n"),
		},
		"prometheus.algo": {
			Data: []byte("\"first-line\nsecond-line\nthird-line\n"),
		},
		"prometheus": {},
	}

	got := treeprint.Generate(fsys)

	if diff := cmp.Diff(got, string(want)); diff != "" {
		t.Log(got)
		t.Errorf("output not matching; (-got, +want):\n %s", diff)
	}
}
