package mdblock

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/diff"

	"gopkg.in/akavel/vfmd.v1"
	"gopkg.in/akavel/vfmd.v1/md"
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
func mkrun(line int, s string) md.Run { return md.Run{line, []byte(s)} }
func bb(s string) []byte              { return []byte(s) }

var newApproach_flatOutput = []md.Tag{
	md.QuoteBlock{Raw: md.Raw{
		mkrun(0, "> * some text **specifically *interesting*** for us.\n"),
		mkrun(1, "> * ## Hello, **[new](http://vfmd.org)** _world._\n"),
		mkrun(2, "![](https://upload.wikimedia.org/wikipedia/commons/1/12/Wikipedia.png)"),
	}},
	md.UnorderedListBlock{Raw: md.Raw{
		mkrun(0, "* some text **specifically *interesting*** for us.\n"),
		mkrun(1, "* ## Hello, **[new](http://vfmd.org)** _world._\n"),
		mkrun(2, "![](https://upload.wikimedia.org/wikipedia/commons/1/12/Wikipedia.png)"),
	}, Starter: mkrun(0, "* "),
	},
	md.ItemBlock{Raw: md.Raw{
		mkrun(0, "* some text **specifically *interesting*** for us.\n"),
	}},
	md.ParagraphBlock{Raw: md.Raw{
		mkrun(0, "some text **specifically *interesting*** for us.\n"),
	}},
	md.Prose{mkrun(-1, "some text ")},
	md.Emphasis{Level: 2},
	md.Prose{mkrun(-1, "specifically ")},
	md.Emphasis{Level: 1},
	md.Prose{mkrun(-1, "interesting")},
	md.End{}, // Emph
	md.End{}, // Emph
	md.Prose{mkrun(-1, " for us.")},
	md.End{}, // Para
	md.End{}, // Item
	md.ItemBlock{Raw: md.Raw{
		mkrun(1, "* ## Hello, **[new](http://vfmd.org)** _world._\n"),
		mkrun(2, "![](https://upload.wikimedia.org/wikipedia/commons/1/12/Wikipedia.png)"),
	}},
	md.AtxHeaderBlock{Level: 2, Raw: md.Raw{
		mkrun(1, "## Hello, **[new](http://vfmd.org)** _world._\n"),
	}},
	md.Prose{mkrun(-1, "Hello, ")},
	md.Emphasis{Level: 2},
	md.Link{
		URL:    "http://vfmd.org",
		RawEnd: md.Raw{mkrun(-1, "](http://vfmd.org)")},
	},
	md.Prose{mkrun(-1, "new")},
	md.End{}, // Link
	md.End{}, // Emph
	md.Prose{mkrun(-1, " ")},
	md.Emphasis{Level: 1},
	md.Prose{mkrun(-1, "world.")},
	md.End{}, // Emph
	md.End{}, // Atx
	md.ParagraphBlock{Raw: md.Raw{
		mkrun(2, "![](https://upload.wikimedia.org/wikipedia/commons/1/12/Wikipedia.png)"),
	}},
	md.Image{
		URL:    "https://upload.wikimedia.org/wikipedia/commons/1/12/Wikipedia.png",
		RawEnd: md.Raw{mkrun(-1, "](https://upload.wikimedia.org/wikipedia/commons/1/12/Wikipedia.png)")},
	},
	md.End{}, // Image
	md.End{}, // Para
	md.End{}, // Item
	md.End{}, // List
	md.End{}, // Quote
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
	result, err := QuickParse(bytes.NewReader(prep), BlocksAndSpans, nil, nil)
	if err != nil {
		test.Fatal(err)
	}
	expected := newApproach_flatOutput
	if !reflect.DeepEqual(result, expected) {
		// TODO(akavel): spew.Dump?
		test.Errorf("expected vs. got DIFF:\n%s",
			diff.Diff(spew.Sdump(expected), spew.Sdump(result)))
	}
}
