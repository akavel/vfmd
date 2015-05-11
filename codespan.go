package vfmd

type CodeSpanDetector struct {
	Within bool
	// number of backtick (`) characters:
	here    int
	opening int
}

func (c *CodeSpanDetector) WriteByte(b byte) error {
	if b == '`' {
		// TODO(akavel): properly handle escaped backticks
		c.Within = true
		c.here++
		return nil
	}
	if c.here == 0 {
		return nil
	}
	switch c.opening {
	case 0:
		c.opening = c.here
	case c.here:
		c.opening = 0
		c.Within = false
	}
	c.here = 0
	return nil
}
