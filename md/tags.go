package md // import "gopkg.in/akavel/vfmd.v0/md"

type Link struct{ ReferenceID, URL, Title string }
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
