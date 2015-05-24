package block

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrCannotDetectBlock = errors.New("vfmd: no block detector matched line")
)

type Block struct {
	Detector
	First, Last int // line numbers
}

type Splitter struct {
	Detectors []Detector
	Blocks    []Block
	current   Block
	window    []Line
}

func (s *Splitter) WriteLine(line []byte) error {
	// TODO(mateuszc): add marker to disallow calls if previous call ended with error
	if s.Detectors == nil {
		s.Detectors = DefaultDetectors
	}
	s.window = append(s.window, line)

	switch {
	case s.current.Detector == nil && len(s.window) == 1:
		// If not in a detected block, we must wait till we have two
		// lines
		return nil
	case s.current.Detector == nil:
		// Let's detect a block!
		dets := cloneSlice(s.Detectors)
		var consume, pause int
		for _, d := range dets {
			consume, pause = d.Detect(s.window[0], s.window[1])
			if consume+pause > 0 {
				s.current.Detector = d
				break
			}
		}
		switch {
		case s.current.Detector == nil:
			return ErrCannotDetectBlock
		case consume < 0 || pause < 0 || consume+pause > 2:
			return fmt.Errorf("vfmd: %T.Detect() broke block.Detector contract: must return one of: 0,0; 0,1; 1,0; 1,1; 0,2; 2,0; got: %d,%d",
				s.current.Detector, consume, pause)
		}
		s.current.Last = s.current.First - 1 + consume
		s.window = s.window[consume:]
		rest := s.window[pause:]
		if len(rest) > 0 {
			assert(len(rest) == 1, len(rest), rest)
			s.window = s.window[:pause]
			return s.WriteLine(rest[0])
		}
		return nil
	default:
		n := len(s.window)
		consume, pause := s.current.Continue(s.window[:n-1], s.window[n-1])
		if consume < 0 || pause < 0 || consume+pause > len(s.window) || (pause != 0 && consume+pause != len(s.window)) {
			return fmt.Errorf("vfmd: %T.Continue() broke block.Continue contract: got: %d,%d",
				s.current.Detector, consume, pause)
		}
		s.current.Last += consume
		switch {
		case consume+pause < len(s.window):
			s.emitBlock()
			return s.retry(s.window[consume:])
		default:
			s.window = s.window[consume:]
			return nil
		}
	}
}

func (s *Splitter) Close() error {
	// TODO(mateuszc): add marker to detect multiple closes
	if len(s.window) == 0 {
		if s.current.Detector != nil {
			s.emitBlock()
		}
		return nil
	}
	return s.WriteLine(nil)
	// assert(s.current.Detector != nil, s.current.Detector, s.window)
	// consume, pause := s.current.Continue(s.window, nil)
	// if consume <= 0 || pause < 0 || consume+pause > len(s.window) {
	// 	return fmt.Errorf("vfmd: %T.Continue() broke block.Continue contract: got: %d,%d",
	// 		s.current.Detector, consume, pause)
	// }
	// s.current.Last += consume
	// s.emitBlock()
	// if consume == len(s.window) {
	// 	s.window = nil
	// 	return nil
	// }
	// s.retry(s.window[consume:])
	// return s.Close()
}

func (s *Splitter) emitBlock() {
	s.Blocks = append(s.Blocks, s.current)
	s.current.Detector = nil
	s.current.First = s.current.Last + 1
}

func (s *Splitter) retry(lines []Line) error {
	s.window = nil
	for _, l := range lines {
		// TODO(mateuszc): kinda risky, may potentially exhaust stack?
		err := s.WriteLine(l)
		if err != nil {
			return err
		}
	}
	return nil
}

func cloneSlice(src []Detector) []Detector {
	dst := make([]Detector, len(src))
	for i := range src {
		s := reflect.ValueOf(src[i])
		data := reflect.Indirect(s)
		clone := reflect.New(data.Type())
		clone.Elem().Set(data)
		if s.Type().Kind() == reflect.Ptr {
			// i.e. was: src[i] = &MyStruct{}
			dst[i] = clone.Interface().(Detector)
		} else {
			// i.e. was: src[i] = MyStruct{}
			dst[i] = clone.Elem().Interface().(Detector)
		}
	}
	return dst
}

func assert(condition bool, notes ...interface{}) {
	if !condition {
		panic("assertion failed; values: " + fmt.Sprintln(notes...))
	}
}
