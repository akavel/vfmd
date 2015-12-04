package block_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/diff"

	"gopkg.in/akavel/vfmd.v0"
	. "gopkg.in/akavel/vfmd.v0/newblock"
	"gopkg.in/akavel/vfmd.v0/span"
)

/*
TODO(akavel): new approach:

type Region []Run
type Run struct {
	Line int
	Bytes []byte // with Line, allows to find out position in line
}

type Prose Region
func (Prose) GetProse() Region
// + "xml-like": self-closer (no End{} to find), or not (must find matching End{})

plus see below:
*/
func mkrun(line int, s string) Run { return Run{line, []byte(s)} }

var newApproach_flatOutput = []Tag{
	Quote{},
	UnorderedList{},
	Item{},
	Paragraph{InQuote: true, InList: true},
	Prose{mkrun(0, "some text ")},
	span.EmphasisBegin{Level: 2},
	Prose{mkrun(0, "specifically ")},
	span.EmphasisBegin{Level: 1},
	Prose{mkrun(0, "interesting")},
	End{}, // Emph
	End{}, // Emph
	Prose{mkrun(0, " for us.")},
	End{}, // Para
	End{}, // Item
	Item{},
	AtxHeader{Level: 2},
	Prose{mkrun(1, "Hello, ")},
	&span.EmphasisBegin{Level: 2},
	&span.LinkBegin{},
	Prose{mkrun(1, "new")},
	End{}, // Link
	End{}, // Emph
	span.EmphasisBegin{Level: 1},
	Prose{mkrun(1, " world.")},
	End{}, // Emph
	End{}, // Atx
	Paragraph{},
	span.Image{},
	// no End, Image is self-closing!
	End{}, // Para
	End{}, // Item
	End{}, // List
	End{}, // Quote
}

var newApproach_outputSketch = `[]interface{}{
&Quote{},
&UL{RawStarter: Region{...}},
&Item{},
&Para{InQuote:true, InList:true},
 &Prose{ /* "some text " */ },
 &Emph{Level:2},
  &Prose{ /* "specifically " */ },
  &Emph{Level:1},
   &Prose{ /* "interesting" */ },
  End{}, // Emph
 End{}, // Emph
 &Prose{ /* " for us." */ },
End{}, // Para
End{}, // Item
&Item{},
&Atx{Level:2},
 &Prose{ /* "Hello, " */ },
 &Emph{Level:2},
  &Link{RawURL: Region{...}},
   &Prose{ /* "new" */ },
  End{}, // Link
 End{}, // Emph
 &Emph{Level:1},
  &Prose{ /* " world." */ },
 End{}, // Emph
End{}, // Atx
Para{},
 &Image{RawURL: Region{...}, RawTitle: Region{...}, RawAlt: Region{...}},
 // no End, Image is self-closing!
End{}, // Para
End{}, // Item
End{}, // UL
End{}, // Quote
}`

var newApproach_input = `> * some text **specifically *interesting*** for us.
> * ## Hello, **[new](http://vfmd.org)** _world._
![](https://upload.wikimedia.org/wikipedia/commons/1/12/Wikipedia.png)`

func init() {
	spew.Config.Indent = "  "
}

func TestNewApproach(test *testing.T) {
	prep, _ := vfmd.QuickPrep(strings.NewReader(newApproach_input))
	result, err := QuickParse(bytes.NewReader(prep), BlocksAndSpans, nil)
	if err != nil {
		test.Fatal(err)
	}
	// FIXME(akavel): all Tags should retain field .Raw (type Region) containing injected lines/spans (with trailing newlines where appropriate)
	expected := newApproach_flatOutput
	if !reflect.DeepEqual(result, expected) {
		// TODO(akavel): spew.Dump?
		test.Errorf("expected vs. got DIFF:\n%s",
			diff.Diff(spew.Sdump(expected), spew.Sdump(result)))
	}
}
