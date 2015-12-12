package md // import "gopkg.in/akavel/vfmd.v0/md"

type Link struct {
	ReferenceID, URL, Title string
	RawEnd                  Raw
}
type AutomaticLink struct{ URL, Text string }
type Emphasis struct{ Level int }
type Code struct{ Code []byte }
type Image struct {
	ReferenceID string
	URL         string
	Title       string
	AltText     []byte
}
type End struct{}

type NullBlock struct {
	Raw
}
type SetextHeaderBlock struct {
	Level int
	Raw
}
type CodeBlock struct {
	Prose
	Raw
}
type AtxHeaderBlock struct {
	Level int
	Raw
}
type QuoteBlock struct {
	Raw
}
type HorizontalRuleBlock struct {
	Raw
}
type UnorderedListBlock struct {
	Starter Run
	Raw
}
type OrderedListBlock struct {
	Starter Run
	Raw
}
type ItemBlock struct {
	Raw
}
type ParagraphBlock struct {
	Raw
}
type ReferenceResolutionBlock struct {
	ReferenceID string
	URL, Title  string
	Raw
}

type Raw Region

func (r Raw) GetRaw() Region { return Region(r) }
