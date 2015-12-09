package block // import "gopkg.in/akavel/vfmd.v0/block"

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/span"
)

// func unstack() {
// 	pc, _, _, _ := runtime.Caller(2)
// 	for i, j := 2, 0; i < 100; i++ {
// 		ipc, _, _, _ := runtime.Caller(i)
// 		if ipc == pc {
// 			j++
// 		}
// 		if j == 10 {
// 			panic("stack too high")
// 		}
// 	}
// }
func unstack() {}

type Mode int

const (
	// TODO(akavel): make sure they're ordered & named as I wanted
	BlocksAndSpans Mode = iota
	BlocksOnly
	TopBlocks
)

type Context interface {
	GetMode() Mode
	GetDetectors() Detectors
	GetSpanDetectors() []span.Detector
	Emit(md.Tag)
}

type defaultContext struct {
	mode          Mode
	tags          []md.Tag
	detectors     Detectors
	spanDetectors []span.Detector
}

func (c *defaultContext) GetMode() Mode                     { return c.mode }
func (c *defaultContext) GetDetectors() Detectors           { return c.detectors }
func (c *defaultContext) GetSpanDetectors() []span.Detector { return c.spanDetectors }
func (c *defaultContext) Emit(tag md.Tag)                   { c.tags = append(c.tags, tag) }

type Parser struct {
	Context

	start   *Line
	handler Handler
}

// func (p *Parser) Emit(tag Tag) { unstack(); p.Context.Emit(tag) }
func (p *Parser) Close() error { return p.WriteLine(Line{}) }
func (p *Parser) WriteLine(line Line) error {
	unstack()

	// Continue previous block if appropriate.
	if p.handler != nil {
		// NOTE(akavel): assert(p.start==nil)
		// fmt.Printf("...handle? %d %q\n", line.Line, string(line.Bytes))
		consumed, err := p.handler.Handle(line, p)
		// fmt.Printf("...handled %v\n", consumed)
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
	p.handler = p.GetDetectors().Find(*p.start, line)
	if p.handler == nil {
		// TODO(akavel): return error object with line number and contents
		return fmt.Errorf("vfmd: no block detector matched line %d: %q", p.start.Line, string(p.start.Bytes))
	}
	// fmt.Printf(".:.handle? %d %q\n", p.start.Line, string(p.start.Bytes))
	consumed, err := p.handler.Handle(*p.start, p)
	// fmt.Printf(".:.handled %v\n", consumed)
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
func QuickParse(r io.Reader, mode Mode, detectors Detectors, spanDetectors []span.Detector) ([]md.Tag, error) {
	scan := bufio.NewScanner(r)
	scan.Split(splitKeepingEOLs)
	if detectors == nil {
		detectors = DefaultDetectors
	}
	if spanDetectors == nil {
		spanDetectors = span.DefaultDetectors
	}
	context := &defaultContext{
		mode:          mode,
		detectors:     detectors,
		spanDetectors: spanDetectors,
	}
	parser := Parser{
		Context: context,
	}
	for i := 0; scan.Scan(); i++ {
		// fmt.Print(scan.Text())
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
type Line md.Run

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
	DetectorFunc(DetectNull),
	// &ReferenceResolution{},
	DetectorFunc(DetectSetextHeader),
	DetectorFunc(DetectCode),
	DetectorFunc(DetectAtxHeader),
	DetectorFunc(DetectQuote),
	DetectorFunc(DetectHorizontalRule),
	DetectorFunc(DetectUnorderedList),
	DetectorFunc(DetectOrderedList),
	ParagraphDetector{},
}

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
	var err error
	if parser != nil {
		err = parser.Close()
	}
	ctx.Emit(md.End{})
	return false, err
}
func end2(parser *Parser, ctx Context) (bool, error) {
	var err error
	if parser != nil {
		err = parser.Close()
		ctx.Emit(md.End{})
	}
	ctx.Emit(md.End{})
	return false, err
}
func pass(parser *Parser, next Line, bytes []byte) (bool, error) {
	if parser != nil {
		return true, parser.WriteLine(Line{next.Line, bytes})
	} else {
		return true, nil
	}
}
func trimLeftN(s []byte, cutset string, nmax int) []byte {
	for nmax > 0 && len(s) > 0 && strings.IndexByte(cutset, s[0]) != -1 {
		nmax--
		s = s[1:]
	}
	return s
}

func parseSpans(region md.Raw, ctx Context) {
	if ctx.GetMode() != BlocksAndSpans {
		return
	}
	// FIXME(akavel): parse the spans correctly w.r.t. region Run boundaries and collect proper Run.Line values
	buf := []byte{}
	for _, run := range region {
		buf = append(buf, run.Bytes...) // FIXME(akavel): quick & dirty & foul prototyping hack
	}
	spans := span.Parse(buf, ctx.GetSpanDetectors())
	for _, span := range spans {
		ctx.Emit(span)
	}
}
