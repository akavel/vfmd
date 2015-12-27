package mdutils

import (
	"io/ioutil"
	"reflect"
	"testing"

	"gopkg.in/akavel/vfmd.v0/md"
)

var (
	buf1 = []byte("Ala ma kota, a kot ma Alę;")
	buf2 = []byte("Ona go kocha, a on ją wcale.")
	r1   = md.Region{
		{0, buf1[:3]},
		{0, buf1[3:]},
		{1, buf2[:6]},
		{1, buf2[6:]},
	}
)

func TestCopy(test *testing.T) {
	r := r1
	c := Copy(r)
	if !reflect.DeepEqual(r, c) {
		test.Fatalf("want:\n%#v\ngot:\n%#v", r, c)
	}
	for i := range r {
		// Bytes buffers must have the same underlying arrays
		if &r[i].Bytes[0] != &c[i].Bytes[0] {
			test.Errorf("&r[%d].Bytes[0] not equal for c", i)
		}
		// Runs in the Region must be different objects
		if &r[i] == &c[i] {
			test.Errorf("&r[%d] is the same for c", i)
		}
	}
}

func TestRegionReader(test *testing.T) {
	cases := []struct {
		r        md.Region
		expected string
	}{
		{r1, "Ala ma kota, a kot ma Alę;Ona go kocha, a on ją wcale."},
	}
	for _, c := range cases {
		r := Copy(c.r)
		rr := regionReader{r: r}
		all, err := ioutil.ReadAll(&rr)
		if err != nil {
			test.Errorf("case %q error: %v", c.expected, err)
		}
		if string(all) != c.expected {
			test.Errorf("want:\n%q\ngot:\n%q", c.expected, string(all))
		}
	}
}

func TestSkip(test *testing.T) {
	r := Copy(r1)
	Skip(&r, 10)
	expected := md.Region{
		{0, buf1[10:]},
		{1, buf2[:6]},
		{1, buf2[6:]},
	}
	if !reflect.DeepEqual(expected, r) {
		test.Fatalf("want:\n%#v\ngot:\n%#v", expected, r)
	}
	for i := range r {
		if !sameArray(r[i].Bytes, expected[i].Bytes) {
			test.Errorf("r[%d].Bytes not same as expected", i)
		}
	}
}

func TestLimit(test *testing.T) {
	r := Copy(r1)
	Limit(&r, 10)
	expected := md.Region{
		{0, buf1[:3]},
		{0, buf1[3:10]},
	}
	if !reflect.DeepEqual(expected, r) {
		test.Fatalf("want:\n%#v\ngot:\n%#v", expected, r)
	}
	for i := range r {
		if !sameArray(r[i].Bytes, expected[i].Bytes) {
			test.Errorf("r[%d].Bytes not same as expected", i)
		}
	}
}

func sameArray(a, b []byte) bool {
	return &a[:cap(a)][cap(a)-1] == &b[:cap(b)][cap(b)-1]
}
