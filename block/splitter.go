package block

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrCannotDetectBlock = errors.New("vfmd: no block detector matched line")
)

type Detection struct {
	Detector
	First, Last int // line numbers
}

type Splitter struct {
	Detectors []Detector
	Detected  []Detection
	consumed  int // number of lines consumed

	block  Detector
	paused []Line
	start  int // nuber of line where current block started
}

func (s *Splitter) WriteLine(line []byte) error {
	if s.Detectors == nil {
		s.Detectors = DefaultDetectors
	}

	switch {
	case s.block == nil && len(s.paused) < 2:
		// If not in a detected block, we must wait till we have two
		// lines
		s.paused = append(s.paused, line)
		return nil
	case s.block == nil:
		// Let's detect a block!
		dets := cloneSlice(s.Detectors)
		var consume, pause int
		for _, d := range dets {
			consume, pause = d.Detect(s.paused[0], s.paused[1])
			if consume+pause > 0 {
				s.block = d
				break
			}
		}
		switch {
		case s.block == nil:
			return ErrCannotDetectBlock
		case consume < 0 || pause < 0 || consume+pause > 2:
			return fmt.Errorf("vfmd: %T.Detect() broke block.Detector contract: must return one of: 0,0; 0,1; 1,0; 1,1; 0,2; 2,0; got: %d,%d",
				s.block, consume, pause)
		}
		s.start = s.consumed
		s.consumed += consume
		s.paused = s.paused[consume:]
		if pause == 0 && len(s.paused) > 0 {
			assert(len(s.paused) == 1, len(s.paused), s.paused)
			retry := s.paused[0]
			s.paused = nil
			return s.WriteLine(retry)
		}
		return nil
	default:
		consume, pause := s.block.Continue(s.paused, line)
		assert(consume >= 0 && pause >= 0, consume, pause)
		assert(consume+pause <= len(s.paused)+1, consume, pause, len(s.paused))
		switch {
		case consume <= len(s.paused):
			s.consumed += consume
			s.Detected = append(s.Detected, Detection{
				Detector: s.block,
				First:    s.start,
				Last:     s.consumed - 1,
			})
			rest := append(s.paused[consume:], line)
			s.paused = nil
			for _, retry := range rest {
				// TODO(mateuszc): kinda risky, may potentially exhaust stack?
				err := s.WriteLine(retry)
				if err != nil {
					return err
				}
			}
			return nil
		default:
			s.paused = nil
			s.consumed += consume
			return nil
		}
	}
}

func (s *Splitter) Close() error {
	// TODO(akavel): NIY
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
