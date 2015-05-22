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
	paused    []Line
}

func (s *Splitter) WriteLine(line []byte) error {
	if s.Detectors == nil {
		s.Detectors = DefaultDetectors
	}

	switch {
	case s.current.Detector == nil && len(s.paused) < 2:
		// If not in a detected block, we must wait till we have two
		// lines
		s.paused = append(s.paused, line)
		return nil
	case s.current.Detector == nil:
		// Let's detect a block!
		dets := cloneSlice(s.Detectors)
		var consume, pause int
		for _, d := range dets {
			consume, pause = d.Detect(s.paused[0], s.paused[1])
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
		s.current.Last += consume
		s.paused = s.paused[consume:]
		if pause == 0 && len(s.paused) > 0 {
			assert(len(s.paused) == 1, len(s.paused), s.paused)
			retry := s.paused[0]
			s.paused = nil
			return s.WriteLine(retry)
		}
		return nil
	default:
		consume, pause := s.current.Continue(s.paused, line)
		assert(consume >= 0 && pause >= 0, consume, pause)
		assert(consume+pause <= len(s.paused)+1, consume, pause, len(s.paused))
		s.current.Last += consume
		switch {
		case consume <= len(s.paused):
			s.emitBlock()
			return s.retry(append(s.paused[consume:], line))
		default:
			s.paused = nil
			return nil
		}
	}
}

func (s *Splitter) Close() error {
	if len(s.paused) == 0 {
		if s.current.Detector != nil {
			s.emitBlock()
		}
		return nil
	}
	assert(s.current.Detector != nil, s.current.Detector)
	consume, pause := s.current.Continue(s.paused, nil)
	assert(consume+pause <= len(s.paused), consume, pause, len(s.paused))
	s.current.Last += consume
	s.emitBlock()
	if consume == len(s.paused) {
		s.paused = nil
		return nil
	}
	s.retry(s.paused[consume:])
	return s.Close()
}

func (s *Splitter) emitBlock() {
	s.Blocks = append(s.Blocks, s.current)
	s.current.Detector = nil
	s.current.First = s.current.Last + 1
	s.current.Last = s.current.First - 1
}

func (s *Splitter) retry(lines []Line) error {
	s.paused = nil
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
