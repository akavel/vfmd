package block

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type Mode int

const (
	// TODO(akavel): make sure they're ordered & named as I wanted
	BlocksAndSpans Mode = iota
	BlocksOnly
	TopBlocks
)

type Tag interface{}
type End struct{}

type Region []Run

// TODO(akavel): rename Run to something prettier?
type Run struct {
	Line  int
	Bytes []byte // with Line, allows to find out position in line
}

// func (r Run) String() string { return string(r.Bytes) }

type Prose Region

func (p Prose) Prose() Region { return Region(p) }

type Proser interface {
	Prose() Region
}

type Context interface {
	Mode() Mode
	Emit(Tag)
}

// Important: r must be pre-processed with vfmd.QuickPrep or vfmd.Preprocessor
func QuickParse(r io.Reader, mode Mode, detectors Detectors) ([]Tag, error) {
	// TODO(akavel): extract below block to a struct
	scan := bufio.NewScanner(r)
	scan.Split(splitKeepingEOLs)
	i := 0
	scanLine := func() *Line {
		if !scan.Scan() {
			return nil
		}
		i++
		return &Line{
			Line: i - 1,
			// Copy the line contents so that scan.Scan() doesn't invalidate it
			Bytes: append([]byte(nil), scan.Bytes()...),
		}
	}

	// Preparations.
	if detectors == nil {
		detectors = DefaultDetectors
	}
	context := &context{mode: mode}
	var handler Handler
	var line *Line
	for {
		if line == nil {
			line = scanLine()
			if line == nil {
				break
			}
		}
		if handler != nil && handler.Handle(line, context) {
			line = nil
			continue
		}
		handler = nil

		// Fetch second line and detect a block.
		second := scanLine()
		handler = detectors.Find(line, second)
		if handler == nil {
			// TODO(akavel): return error object with line number and contents
			return nil, fmt.Errorf("vfmd: no block detector matched line %d: %q", line.Line, string(line.Bytes))
		}
		if !handler.Handle(line, context) {
			return nil, fmt.Errorf("vfmd: detector %T failed to handle first line %d: %q", handler, line.Line, string(line.Bytes))
		}
		if second == nil {
			break
		}
		line = second
	}
	if scan.Err() != nil {
		return nil, scan.Err()
	}
	if handler != nil {
		handler.Handle(nil, context)
	}
	return context.tags, nil
}

// Line is a Run that may have at most one '\n', as last byte
type Line Run

func (line Line) isBlank() bool {
	return len(bytes.Trim(line.Bytes, " \t\n")) == 0
}
func (line Line) hasNonSpaceInPrefix(n int) bool {
	bs := line.Bytes
	for i := 0; i < n && i < len(bs) && bs[i] != '\n'; i++ {
		if bs[i] != ' ' {
			return true
		}
	}
	return false
}
func (line Line) hasFourSpacePrefix() bool {
	return bytes.HasPrefix(line.Bytes, []byte("    "))
}

type Detector interface {
	Detect(first, second *Line, detectors Detectors) Handler
}
type Handler interface {
	Handle(*Line, Context) (consumed bool)
}

type DetectorFunc func(first, second *Line, detectors Detectors) Handler
type HandlerFunc func(*Line, Context) (consumed bool)

func (f DetectorFunc) Detect(first, second *Line, detectors Detectors) Handler {
	return f(first, second, detectors)
}
func (f HandlerFunc) Handle(line *Line, context Context) bool {
	return f(line, context)
}

type context struct {
	mode Mode
	tags []Tag
}

func (c *context) Mode() Mode   { return c.mode }
func (c *context) Emit(tag Tag) { c.tags = append(c.tags, tag) }

func splitKeepingEOLs(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, c := range data {
		if c == '\n' {
			return i + 1, data[:i+1], nil
		}
	}
	switch {
	case !atEOF:
		return 0, nil, nil
	case len(data) > 0:
		return len(data), data, nil
	default:
		return 0, nil, io.EOF
	}
}

type Detectors []Detector

// DefaultDetectors contains the list of default detectors in order in which
// they should be normally applied.
// FIXME(akavel): fill DefaultDetectors
var DefaultDetectors = Detectors{
	// Null{},
	// &ReferenceResolution{},
	// &SetextHeader{},
	// &Code{},
	// &AtxHeader{},
	&Quote{},
	// HorizontalRule{},
	// &UnorderedList{},
	// &OrderedList{},
	// &Paragraph{},
}

func (ds Detectors) Find(first, second *Line) Handler {
	for _, d := range ds {
		handler := d.Detect(first, second, ds)
		if handler != nil {
			return handler
		}
	}
	return nil
}
