package block

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
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
	GetMode() Mode
	Emit(Tag)
}

type defaultContext struct {
	mode Mode
	tags []Tag
}

func (c *defaultContext) GetMode() Mode { return c.mode }
func (c *defaultContext) Emit(tag Tag)  { c.tags = append(c.tags, tag) }

type Parser struct {
	Detectors Detectors
	Context

	start   *Line
	handler Handler
}

func (p *Parser) Close() error { return p.WriteLine(Line{}) }
func (p *Parser) WriteLine(line Line) error {
	if p.Detectors == nil {
		p.Detectors = *defaultDetectors
	}

	// Continue previous block if appropriate.
	if p.handler != nil {
		// NOTE(akavel): assert(p.start==nil)
		consumed, err := p.handler.Handle(line, p)
		if err != nil {
			return err
		}
		if line.EOF() {
			return nil
		}
		if consumed {
			return nil
		}
	}
	p.handler = nil

	// New block needs 2 lines for detection.
	if p.start == nil {
		if line.EOF() {
			return nil
		}
		p.start = &line
		return nil
	}
	p.handler = p.Detectors.Find(*p.start, line)
	if p.handler == nil {
		// TODO(akavel): return error object with line number and contents
		return fmt.Errorf("vfmd: no block detector matched line %d: %q", p.start.Line, string(p.start.Bytes))
	}
	consumed, err := p.handler.Handle(*p.start, p)
	if err != nil {
		return err
	}
	if !consumed {
		return fmt.Errorf("vfmd: detector %T failed to handle first line %d: %q", p.handler, p.start.Line, string(p.start.Bytes))
	}
	p.start = nil
	return p.WriteLine(line)
}

// Important: r must be pre-processed with vfmd.QuickPrep or vfmd.Preprocessor
func QuickParse(r io.Reader, mode Mode, detectors Detectors) ([]Tag, error) {
	scan := bufio.NewScanner(r)
	scan.Split(splitKeepingEOLs)
	context := &defaultContext{
		mode: mode,
	}
	parser := Parser{
		Context:   context,
		Detectors: detectors,
	}
	for i := 0; scan.Scan(); i++ {
		err := parser.WriteLine(Line{
			Line: i,
			// Copy the line contents so that scan.Scan() doesn't invalidate it
			Bytes: append([]byte(nil), scan.Bytes()...),
		})
		if err != nil {
			return nil, err
		}
	}
	if scan.Err() != nil {
		return nil, scan.Err()
	}
	err := parser.Close()
	if err != nil {
		return nil, err
	}
	return context.tags, nil
}

// Line is a Run that may have at most one '\n', as last byte
type Line Run

func (line Line) EOF() bool { return line.Bytes == nil }
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
	Detect(first, second Line, detectors Detectors) Handler
}
type Handler interface {
	Handle(Line, Context) (consumed bool, err error)
}

type DetectorFunc func(first, second Line, detectors Detectors) Handler
type HandlerFunc func(Line, Context) (consumed bool, err error)

func (f DetectorFunc) Detect(first, second Line, detectors Detectors) Handler {
	return f(first, second, detectors)
}
func (f HandlerFunc) Handle(line Line, context Context) (bool, error) {
	return f(line, context)
}

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
	DetectorFunc(DetectQuote),
	// HorizontalRule{},
	DetectorFunc(DetectUnorderedList),
	// &OrderedList{},
	// &Paragraph{},
}

// defaultDetectors helps break initialization loop for elements of
// DefaultDetectors referencing DefaultDetectors.
var defaultDetectors *Detectors

func init() { defaultDetectors = &DefaultDetectors }

func (ds Detectors) Find(first, second Line) Handler {
	for _, d := range ds {
		handler := d.Detect(first, second, ds)
		if handler != nil {
			return handler
		}
	}
	return nil
}

func end(parser *Parser, ctx Context) (bool, error) {
	err := parser.Close()
	ctx.Emit(End{})
	return false, err
}
func end2(parser *Parser, ctx Context) (bool, error) {
	err := parser.Close()
	ctx.Emit(End{})
	ctx.Emit(End{})
	return false, err
}
func pass(parser *Parser, next Line, bytes []byte) (bool, error) {
	return true, parser.WriteLine(Line{next.Line, bytes})
}
func trimLeftN(s []byte, cutset string, nmax int) []byte {
	for nmax > 0 && len(s) > 0 && strings.IndexByte(cutset, s[0]) != -1 {
		nmax--
		s = s[1:]
	}
	return s
}
