package mdblock

import (
	"testing"

	"gopkg.in/akavel/vfmd.v0/md"
)

func TestRefResolution(test *testing.T) {
	region := md.Region{
		mkrun(0, "[ref]: url1\n"),
		mkrun(1, "[link `containing code`]: url2\n"),
	}
	handler := DetectReferenceResolution(Line(region[0]), Line(region[1]), nil)
	if handler == nil {
		test.Fatal("handler==nil")
	}
}
