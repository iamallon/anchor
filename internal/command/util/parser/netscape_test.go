package parser_test

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/loghinalexandru/anchor/internal/command/util/parser"
	"github.com/virtualtam/netscape-go/v2"
)

type TraversalTest struct {
	name    string
	content []string
}

func TestTraverseNodeWithError(t *testing.T) {
	t.Parallel()

	err := parser.TraverseNode("", nil, netscape.Folder{
		Bookmarks: []netscape.Bookmark{
			{
				Title: "invalid labels bookmark",
			},
		},
	})

	if err == nil {
		t.Error("missing expected error")
	}
}

func TestTraverseNodeEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	input, err := os.ReadFile("testdata/empty.input")
	if err != nil {
		t.Fatalf("unexpected error when reading input file; got %s", err)
	}

	doc, err := netscape.Unmarshal(input)
	if err != nil {
		t.Fatalf("unexpected error when parsing input file; got %s", err)
	}

	err = parser.TraverseNode(dir, nil, doc.Root)
	if err != nil {
		t.Fatalf("unexpected error; got %s", err)
	}

	got, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("unexpected error; got %s", err)
	}

	if len(got) > 0 {
		t.Errorf("result directory is not empty; got %s", got)
	}
}

func TestTraverseNodeComplex(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	want := []TraversalTest{
		{
			name: "root",
			content: []string{
				`"Outlook" "https://outlook.live.com/mail/0/"`,
				`"Gmail" "https://accounts.google.com/b/0/AddMailService"`,
				`"YouTube" "https://youtube.com/"`,
				`"YouTube" "https://youtube.com/"`,
			},
		},
		{
			name: "gan",
			content: []string{
				`"Introduction to GANs with Python and TensorFlow" "https://stackabuse.com/introduction-to-gans-with-python-and-tensorflow/"`,
				`"sklearn.datasets.fetch_lfw_people — scikit-learn 0.24.1 documentation" "https://scikit-learn.org/stable/modules/generated/sklearn.datasets.fetch_lfw_people.html"`,
				`"A Beginner's Guide to Generative Adversarial Networks (GANs) | Pathmind" "https://wiki.pathmind.com/generative-adversarial-network-gan"`,
				`"mnist-gan/gan.py at master · gtoubassi/mnist-gan" "https://github.com/gtoubassi/mnist-gan/blob/master/gan.py"`,
				`"GitHub - soumith/ganhacks: starter from \"How to Train a GAN?\" at NIPS2016" "https://github.com/soumith/ganhacks"`,
			},
		},
		{
			name: "gan.research-papers",
			content: []string{
				`"GAN - 2014 paper" "https://arxiv.org/pdf/1406.2661.pdf"`,
				`"LSGAN.pdf" "https://arxiv.org/pdf/1611.04076.pdf"`,
				`"Internal Covariate Shift.pdf" "https://arxiv.org/pdf/1502.03167.pdf"`,
				`"https://arxiv.org/pdf/1903.06048.pdf" "https://arxiv.org/pdf/1903.06048.pdf"`,
				`"https://arxiv.org/pdf/1802.05957.pdf" "https://arxiv.org/pdf/1802.05957.pdf"`,
			},
		},
		{
			name:    "inlineskating",
			content: []string{`"Joe Atkinson" "https://joe-atkinson.com/"`},
		},
		{
			name: "technicalbooks",
			content: []string{
				`"Test Driven Development: By Example: Beck, Kent: 8601400403228: Amazon.com: Books" "https://www.amazon.com/Test-Driven-Development-Kent-Beck/dp/0321146530?crid=1XOC3LK98ZBQJ&keywords=kent+beck+test+driven+development&qid=1659096793&sprefix=kent+beck,aps,530&sr=8-1&linkCode=sl1&tag=cribbcorne-20&linkId=0f7f1edb0fb222a629ea4f2051ef1a66&language=en_US&ref_=as_li_ss_tl"`,
				`"C4 model for visualising… by Simon Brown [PDF/iPad/Kindle]" "https://leanpub.com/visualising-software-architecture"`,
				`"Refactoring: Improving the Design of Existing Code (2nd Edition) (Addison-Wesley Signature Series (Fowler)): Fowler, Martin: 9780134757599: Amazon.com: Books" "https://www.amazon.com/Refactoring-Improving-Existing-Addison-Wesley-Signature/dp/0134757599/ref=pd_bxgy_img_sccl_1/139-2960292-6658539?pd_rd_w=WSNcL&content-id=amzn1.sym.6ab4eb52-6252-4ca2-a1b9-ad120350253c&pf_rd_p=6ab4eb52-6252-4ca2-a1b9-ad120350253c&pf_rd_r=ST766MEC5CRBP56PJ0GX&pd_rd_wg=ygkdF&pd_rd_r=93a24c09-007c-4384-aee6-2f513aac0ee3&pd_rd_i=0134757599&psc=1"`,
				`"Software Architecture in Practice (SEI Series in Software Engineering): Amazon.co.uk: Bass, Len, Clements, Paul, Kazman, Rick: 9780321815736: Books" "https://www.amazon.co.uk/dp/0321815734/ref=as_li_tl?ie=UTF8&linkCode=gg2&linkId=a35d633edf08483b3a23986c745d0510&creativeASIN=0321815734&tag=ashanin-20&creative=9325&camp=1789"`,
			},
		},
		{
			name: "technicalbooks.architecture",
			content: []string{
				`"The Art of Immutable Architecture: Theory and Practice of Data Management in Distributed Systems: Perry, Michael L.: 9781484259542: Amazon.com: Books" "https://www.amazon.com/Art-Immutable-Architecture-Management-Distributed/dp/1484259548"`,
				`"Patterns of Enterprise Application Architecture: Fowler, Martin: 8601300201672: Amazon.com: Books" "https://www.amazon.com/Patterns-Enterprise-Application-Architecture-Martin/dp/0321127420/ref=sr_1_10?crid=26QDOLIMGA3RB&keywords=software+architecture&qid=1680872862&s=books&sprefix=softwarearchitecture%2Cstripbooks-intl-ship%2C176&sr=1-10"`,
			},
		},
		{
			name: "technicalbooks.architecture.classics",
			content: []string{
				`"Working Effectively with Legacy Code: Feathers, Michael: 8601400968741: Amazon.com: Books" "https://www.amazon.com/Working-Effectively-Legacy-Michael-Feathers/dp/0131177052?keywords=working+effectively+with+legacy+code&qid=1655807854&s=books&sprefix=working+effecti,stripbooks,138&sr=1-1&linkCode=sl1&tag=cribbcorne-20&linkId=0f102683446ee44df1a08149e0575dd1&language=en_US&ref_=as_li_ss_tl"`,
				`"Perfect Software: And Other Illusions about Testing: Gerald M. Weinberg: 9780932633699: Amazon.com: Books" "https://www.amazon.com/Perfect-Software-Other-Illusions-Testing-dp-0932633692/dp/0932633692?_encoding=UTF8&me=&qid=1655806530&linkCode=sl1&tag=cribbcorne-20&linkId=6001f1bb2e483198ecd1a625ae165f88&language=en_US&ref_=as_li_ss_tl"`,
			},
		},
	}

	input, err := os.ReadFile("testdata/complex.input")
	if err != nil {
		t.Fatalf("unexpected error when reading input file; got %s", err)
	}

	doc, err := netscape.Unmarshal(input)
	if err != nil {
		t.Fatalf("unexpected error when parsing input file; got %s", err)
	}

	err = parser.TraverseNode(dir, nil, doc.Root)
	if err != nil {
		t.Fatalf("unexpected error; got %s", err)
	}

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		testCase := slices.IndexFunc(want, func(tt TraversalTest) bool {
			if tt.name == d.Name() {
				return true
			}

			return false
		})

		if testCase == -1 {
			t.Errorf(fmt.Sprintf("could not find file name %q", d.Name()))
			return nil
		}

		fh, err := os.Open(path)
		if err != nil {
			return err
		}

		var idx int
		defer fh.Close()
		got := bufio.NewScanner(fh)
		for got.Scan() {
			if !strings.HasPrefix(got.Text(), want[testCase].content[idx]) {
				t.Fatalf("mismatch content for file %q; (-want +got):\n %s", d.Name(), cmp.Diff(want[testCase].content[idx], got.Text()))
			}
			idx++
		}

		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error; got %s", err)
	}
}
