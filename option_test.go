package easycsv

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"
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

func TestCustomDecoder(t *testing.T) {
	f := bytes.NewBufferString("hello,2010-11-12\nworld,2012-01-02")
	r := NewReader(f, Option{
		Decoders: map[string]interface{}{
			"custom": func(s string) (string, error) { return "[" + s + "]", nil },
			"date": func(s string) (time.Time, error) {
				return time.Parse("2006-01-02", s)
			},
		},
	})
	var msgs []string
	var dates []string
	r.Loop(func(e struct {
		Msg  string    `index:"0" enc:"custom"`
		Date time.Time `index:"1" enc:"date"`
	}) error {
		msgs = append(msgs, e.Msg)
		dates = append(dates, e.Date.Format("2006/1/2"))
		return nil
	})
	if err := r.Done(); err != nil {
		t.Error(err)
	}
	msgExpects := []string{"[hello]", "[world]"}
	dateExpects := []string{"2010/11/12", "2012/1/2"}
	if !reflect.DeepEqual(msgs, msgExpects) {
		t.Errorf("Expected %v but got %v", msgExpects, msgs)
	}
	if !reflect.DeepEqual(dates, dateExpects) {
		t.Errorf("Expected %v but got %v", dateExpects, dates)
	}
}

func TestCustomDecoder_wrongType(t *testing.T) {
	f := bytes.NewBufferString("hello,2010-11-12\nworld,2012-01-02")
	r := NewReader(f, Option{
		Decoders: map[string]interface{}{
			"enc0": nil,
			"enc1": 10,
			"enc2": func() {},
			"enc3": func(n int) (float32, bool) { return 1.0, true },
		},
	})
	r.Loop(func(e struct {
		F0 string `index:"0" enc:"enc0"`
		F1 string `index:"0" enc:"enc1"`
		F2 string `index:"0" enc:"enc2"`
		F3 string `index:"0" enc:"enc3"`
	}) error {
		t.Error("Loop read an entry unexpectedly")
		return nil
	})
	err := r.Done()
	if err == nil {
		t.Errorf("Loop above must fail")
	}
	expectedErrors := []string{
		"Encoding \"enc0\" is not defined",
		"The custom decoder for Encoding \"enc1\" must be a function",
		"The custom decoder for Encoding \"enc2\" must receive an arg, but receives 0 args",
		"The custom decoder for Encoding \"enc2\" must return two values, but returns 0 values",
		"The custom decoder for Encoding \"enc3\" must receive a string, but receives int",
		"The type of field \"F3\" is string, but enc \"enc3\" returns \"float32\"",
		"The second return value of the custom decoder for \"enc3\" must be error",
	}
	if err.Error() != strings.Join(expectedErrors, "\n") {
		t.Errorf("Unexpected errors: %v", err)
	}
}
