package easycsv

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestReadTSV(t *testing.T) {
	f := bytes.NewBufferString("1\t2\n3\t4\n")
	r := NewReader(f, Option{
		Comma: '\t',
	})
	var got [][]int
	if err := r.ReadAll(&got); err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	want := [][]int{{1, 2}, {3, 4}}
	noDiff(t, "ReadAll() with tsv", got, want)
}

func TestSkipComment(t *testing.T) {
	f := bytes.NewBufferString("1,2\n#3,4\n5,6")
	r := NewReader(f, Option{
		Comment: '#',
	})
	var got [][]int
	if err := r.ReadAll(&got); err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	want := [][]int{{1, 2}, {5, 6}}
	noDiff(t, "ReadAll() with Comment", got, want)
}

func TestLazyQuotes(t *testing.T) {
	f := bytes.NewBufferString("1,2,3,\"\"4\",5")
	r := NewReader(f, Option{
		LazyQuotes: true,
		Comment:    '#',
	})
	var got [][]string
	if err := r.ReadAll(&got); err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	want := [][]string{{"1", "2", "3", "\"4", "5"}}
	noDiff(t, "ReadAll() with LazyQuotes", got, want)
}

func TestOptionWithNewReadCloser(t *testing.T) {
	f := &fakeCloser{
		reader: bytes.NewBufferString("1\t2\n3\t4\n"),
	}
	r := NewReadCloser(f, Option{
		Comma: '\t',
	})
	var got [][]int
	if err := r.ReadAll(&got); err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	want := [][]int{{1, 2}, {3, 4}}
	noDiff(t, "results", got, want)
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
	err := r.Loop(func(e struct {
		Msg  string    `index:"0" enc:"custom"`
		Date time.Time `index:"1" enc:"date"`
	}) error {
		msgs = append(msgs, e.Msg)
		dates = append(dates, e.Date.Format("2006/1/2"))
		return nil
	})
	if err != nil {
		t.Fatalf("Loop failed: %v", err)
	}
	msgsWant := []string{"[hello]", "[world]"}
	datesWant := []string{"2010/11/12", "2012/1/2"}
	if diff := cmp.Diff(msgsWant, msgs); diff != "" {
		t.Errorf("mismatch of msgs (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(datesWant, dates); diff != "" {
		t.Errorf("mismatch of dates (-want +got):\n%s", diff)
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
	err := r.Loop(func(e struct {
		F0 string `index:"0" enc:"enc0"`
		F1 string `index:"0" enc:"enc1"`
		F2 string `index:"0" enc:"enc2"`
		F3 string `index:"0" enc:"enc3"`
	}) error {
		t.Error("Loop read an entry unexpectedly")
		return nil
	})
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

func TestTypeDecoders(t *testing.T) {
	f := bytes.NewBufferString("2013-01-02,2010-11-12\n2015-11-19,2012-01-02")
	r := NewReader(f, Option{
		TypeDecoders: map[reflect.Type]interface{}{
			reflect.TypeOf(time.Time{}): func(s string) (time.Time, error) {
				return time.Parse("2006-01-02", s)
			},
		},
	})
	var entry struct {
		Date0 time.Time `index:"0"`
		Date1 time.Time `index:"1"`
	}
	var all []string
	for r.Read(&entry) {
		all = append(all, entry.Date0.Format("2006/1/2"))
		all = append(all, entry.Date1.Format("Jan 2, 2006"))
	}
	if err := r.Done(); err != nil {
		t.Fatalf("Failed to Done: %v", err)
	}
	want := []string{"2013/1/2", "Nov 12, 2010", "2015/11/19", "Jan 2, 2012"}
	noDiff(t, "all", all, want)
}

func TestTypeDecodersWithSlice(t *testing.T) {
	f := bytes.NewBufferString("2013-01-02,2010-11-12\n2015-11-19,2012-01-02")
	r := NewReader(f, Option{
		TypeDecoders: map[reflect.Type]interface{}{
			reflect.TypeOf(time.Time{}): func(s string) (time.Time, error) {
				return time.Parse("2006-01-02", s)
			},
		},
	})
	var row []time.Time
	var all []string
	for r.Read(&row) {
		for _, e := range row {
			all = append(all, e.Format("2006/1/2"))
		}
	}
	if err := r.Done(); err != nil {
		t.Fatalf("Failed to Done: %v", err)
	}
	want := []string{"2013/1/2", "2010/11/12", "2015/11/19", "2012/1/2"}
	noDiff(t, "all", all, want)
}

func TestTypeDecodersErrors(t *testing.T) {
	tests := []struct {
		decoder interface{}
		suberr  string
	}{
		{
			decoder: "decoder",
			suberr:  "must be a function but string",
		}, {
			decoder: func(s string) (int, error) {
				return 0, nil
			},
			suberr: "but returned (int, error)",
		}, {
			decoder: func(s string) time.Time {
				return time.Now()
			},
			suberr: "must receive one arguments and returns two values",
		}, {
			decoder: func(i int) (time.Time, error) {
				return time.Now(), nil
			},
			suberr: "must receive a string as the first arg, but receives int",
		},
	}
	for _, test := range tests {
		f := bytes.NewBufferString("2013-01-02,2010-11-12\n2015-11-19,2012-01-02")
		r := NewReader(f, Option{
			TypeDecoders: map[reflect.Type]interface{}{
				reflect.TypeOf(time.Time{}): test.decoder,
			},
		})
		var row []time.Time
		var all []string
		for r.Read(&row) {
			for _, e := range row {
				all = append(all, e.Format("2006/1/2"))
			}
		}
		err := r.Done()
		if err == nil {
			t.Error("Done() returned nil unexpectedly")
		}
		if !strings.Contains(err.Error(), test.suberr) {
			t.Errorf("%q does not contains %q", err.Error(), test.suberr)
		}
	}
}
