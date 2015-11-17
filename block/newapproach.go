package block

import (
	"bufio"
	"io"
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

// Important: r must be pre-processed with vfmd.QuickPrep or vfmd.Preprocessor
func QuickParse(r io.Reader) ([]Tag, error) {
	// Using old Splitter, find out which lines belong to which blocks.
	// FIXME(akavel): make below two passes into one pass by
	// modernizing/rewriting Splitter
	// FIXME(akavel): resolve EOLs (best keep full lines with EOLs in
	// .Raw [type Region] field of block Tags)
	scan, split, lines := bufio.NewScanner(r), Splitter{}, Region{}
	scan.Split(splitLinesWithEOL)
	for i := 0; scan.Scan(); i++ {
		line := append([]byte(nil), scan.Bytes()...)
		// Store full copy of line for later.
		lines = append(lines, Run{i, line})
		// Pass line without EOL to Splitter (block detector).
		if n := len(line); line[n-1] == '\n' {
			line = line[:n-1]
		}
		err := split.WriteLine(line)
		if err != nil {
			// FIXME(akavel): add info about line number
			return nil, err
		}
	}
	if scan.Err() != nil {
		return nil, scan.Err()
	}
	err := split.Close()
	if err != nil {
		return nil, err
	}

	// Translate results of splitting with old approach to new approach.
	// FIXME(akavel): below code generates only top level of Tags
	output := []Tag{}
	for _, block := range split.Blocks {
		output = append(output,
			block.Detector,
			Region(lines[block.First:block.Last+1]),
			End{})
	}
	return output, nil
}

func splitLinesWithEOL(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
