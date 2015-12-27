package mdutils

import (
	"reflect"
	"testing"

	"gopkg.in/akavel/vfmd.v0/md"
)

func TestCopy(test *testing.T) {
	buf1 := []byte("Ala ma kota, a kot ma Alę;")
	buf2 := []byte("Ona go kocha, a on ją wcale.")
	r := md.Region{
		{0, buf1[:3]},
		{0, buf1[3:]},
		{1, buf2[:6]},
		{1, buf2[6:]},
	}
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
