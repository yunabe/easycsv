package easycsv

import (
	"bytes"
	"reflect"
	"testing"
)

func TestReadTSV(t *testing.T) {
	f := bytes.NewBufferString("1\t2\n3\t4\n")
	r := NewReader(f, Option{
		Comma: '\t',
	})
	var content [][]int
	r.ReadAll(&content)
	if err := r.Done(); err != nil {
		t.Error(err)
	}
	expected := [][]int{{1, 2}, {3, 4}}
	if !reflect.DeepEqual(expected, content) {
		t.Errorf("Expected %v but got %v", expected, content)
	}
}

func TestSkipComment(t *testing.T) {
	f := bytes.NewBufferString("1,2\n#3,4\n5,6")
	r := NewReader(f, Option{
		Comment: '#',
	})
	var content [][]int
	r.ReadAll(&content)
	if err := r.Done(); err != nil {
		t.Error(err)
	}
	expected := [][]int{{1, 2}, {5, 6}}
	if !reflect.DeepEqual(expected, content) {
		t.Errorf("Expected %v but got %v", expected, content)
	}
}

func TestOptionWithNewReadCloser(t *testing.T) {
	f := &fakeCloser{
		reader: bytes.NewBufferString("1\t2\n3\t4\n"),
	}
	r := NewReadCloser(f, Option{
		Comma: '\t',
	})
	var content [][]int
	r.ReadAll(&content)
	if err := r.Done(); err != nil {
		t.Error(err)
	}
	expected := [][]int{{1, 2}, {3, 4}}
	if !reflect.DeepEqual(expected, content) {
		t.Errorf("Expected %v but got %v", expected, content)
	}
	if !f.closed {
		t.Error("f is not closed")
	}
}
