package easycsv

import (
	"bytes"
	"strings"
	"testing"
)

func TestConverterInt(t *testing.T) {
	r := NewReader(bytes.NewBufferString("10,0xff,017"))
	var got []int
	ok := r.Read(&got)
	if !ok {
		t.Fatal("Read returned false unexpectedly")
	}
	want := []int{10, 255, 15}
	noDiff(t, "Read()", got, want)
}

func TestConverterInvalidWithSlice(t *testing.T) {
	r := NewReader(bytes.NewBufferString("hello"))
	var row []int
	ok := r.Read(&row)
	if ok {
		t.Error("Read returned true unexpectedly")
	}
	if err := r.Done(); err == nil || !strings.Contains(err.Error(), "parsing \"hello\"") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestConverterInvalidWithStruct(t *testing.T) {
	r := NewReader(bytes.NewBufferString("hello"))
	var row struct {
		Int int `index:"0"`
	}
	ok := r.Read(&row)
	if ok {
		t.Error("Read returned true unexpectedly")
	}
	if err := r.Done(); err == nil || !strings.Contains(err.Error(), "parsing \"hello\"") {
		t.Errorf("Unexpected error: %v", err)
	}
}
