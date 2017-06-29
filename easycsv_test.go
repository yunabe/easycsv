package easycsv

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)
import "bytes"

func TestLoopNil(t *testing.T) {
	f := bytes.NewReader([]byte(""))
	r := NewReader(f)
	r.Loop(nil)
	err := r.Done()
	if err == nil || !strings.Contains(err.Error(), "must not be nil") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReadNil(t *testing.T) {
	f := bytes.NewReader([]byte(""))
	r := NewReader(f)
	ok := r.Read(nil)
	if ok {
		t.Error("Loop returned true unexpectedly")
		return
	}
	if err := r.Done(); err == nil || !strings.Contains(err.Error(), "must not be nil.") {
		t.Errorf("Unexpected eror: %v", err)
	}
}

type fakeCloser struct {
	reader io.Reader
	err    error
	closed bool
}

func (c *fakeCloser) Close() error {
	c.closed = true
	return c.err
}

func (c *fakeCloser) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func TestCloser(t *testing.T) {
	c := &fakeCloser{}
	r := NewReadCloser(c)
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	if !c.closed {
		t.Error("c is not closed.")
	}
}

func TestCloserWithError(t *testing.T) {
	c := &fakeCloser{
		reader: bytes.NewBufferString(""),
	}
	c.err = errors.New("Close Error")
	r := NewReadCloser(c)
	var unused []string
	if ok := r.Read(&unused); ok {
		t.Errorf("r.Read() must return false for a empty input")
	}
	if err := r.Done(); err != c.err {
		t.Errorf("Unexpected error: %v", err)
	}
	if !c.closed {
		t.Error("c is not closed.")
	}
}

func TestCloserEOFAndError(t *testing.T) {
	c := &fakeCloser{}
	c.err = errors.New("Close Error")
	r := NewReadCloser(c)
	if err := r.Done(); err != c.err {
		t.Errorf("Unexpected error: %v", err)
	}
	if !c.closed {
		t.Error("c is not closed.")
	}
}

func TestCloserDontOverwriteError(t *testing.T) {
	c := &fakeCloser{}
	c.err = errors.New("Close Error")
	r := NewReadCloser(c)
	anotherErr := errors.New("Another error")
	r.err = anotherErr
	if err := r.Done(); err != anotherErr {
		t.Errorf("Unexpected error: %v", err)
	}
	if !c.closed {
		t.Error("c is not closed.")
	}
}

func TestNewReaderFile(t *testing.T) {
	r := NewReaderFile("testing/notexist.csv")
	ok := r.Read(nil)
	if ok {
		t.Error("r.Read returned true unexpectedly")
	}
	if err := r.Done(); err == nil || !strings.Contains(err.Error(), "no such file") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLoop(t *testing.T) {
	f := bytes.NewReader([]byte("10,1.2,alpha\n20,2.3,beta\n30,3.4,gamma"))
	r := NewReader(f)
	var ints []int
	var floats []float32
	var strs []string
	r.Loop(func(e struct {
		Int   int     `index:"0"`
		Float float32 `index:"1"`
		Str   string  `index:"2"`
	}) error {
		ints = append(ints, e.Int)
		floats = append(floats, e.Float)
		strs = append(strs, e.Str)
		return nil
	})
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expectedInt := []int{10, 20, 30}
	expectedFloat := []float32{1.2, 2.3, 3.4}
	expectedStr := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(expectedInt, ints) {
		t.Errorf("Expected %#v but got %#v", expectedInt, ints)
	}
	if !reflect.DeepEqual(expectedFloat, floats) {
		t.Errorf("Expected %#v but got %#v", expectedFloat, floats)
	}
	if !reflect.DeepEqual(expectedStr, strs) {
		t.Errorf("Expected %#v but got %#v", expectedStr, strs)
	}
}

func TestLoopPointer(t *testing.T) {
	f := bytes.NewReader([]byte("10,1.2\n20,2.3\n30,3.4"))
	r := NewReader(f)
	var ints []int
	var floats []float32
	r.Loop(func(e *struct {
		Int   int     `index:"0"`
		Float float32 `index:"1"`
	}) error {
		ints = append(ints, e.Int)
		floats = append(floats, e.Float)
		return nil
	})
	if err := r.Done(); err != nil {
		t.Error(err)
	}
	expectedInt := []int{10, 20, 30}
	expectedFloat := []float32{1.2, 2.3, 3.4}
	if !reflect.DeepEqual(expectedInt, ints) {
		t.Errorf("Unexpected %#v but got %#v", expectedInt, ints)
	}
	if !reflect.DeepEqual(expectedFloat, floats) {
		t.Errorf("Unexpected %#v but got %#v", expectedFloat, floats)
	}
}

func TestLoopWithName(t *testing.T) {
	f := bytes.NewReader([]byte("int,float,str\n10,1.2,alpha\n20,2.3,beta\n30,3.4,gamma"))
	r := NewReader(f)
	var ints []int
	var floats []float32
	var strs []string
	r.Loop(func(e struct {
		Int   int     `name:"int"`
		Float float32 `name:"float"`
		Str   string  `name:"str"`
	}) error {
		ints = append(ints, e.Int)
		floats = append(floats, e.Float)
		strs = append(strs, e.Str)
		return nil
	})
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expectedInt := []int{10, 20, 30}
	expectedFloat := []float32{1.2, 2.3, 3.4}
	expectedStr := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(expectedInt, ints) {
		t.Errorf("Expected %#v but got %#v", expectedInt, ints)
	}
	if !reflect.DeepEqual(expectedFloat, floats) {
		t.Errorf("Expected %#v but got %#v", expectedFloat, floats)
	}
	if !reflect.DeepEqual(expectedStr, strs) {
		t.Errorf("Expected %#v but got %#v", expectedStr, strs)
	}
}

func TestLoopIndexOutOfRange(t *testing.T) {
	f := bytes.NewReader([]byte("10,1.2\n20,2.3"))
	r := NewReader(f)
	r.Loop(func(e struct {
		Int   int     `index:"0"`
		Float float32 `index:"2"`
	}) error {
		t.Error("The callback of Look is invoked unexpectedly")
		return nil
	})
	err := r.Done()
	if err == nil || err.Error() != "Accessed index 2 though the size of the row is 2" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLoopMissingColumn(t *testing.T) {
	f := bytes.NewReader([]byte("a,b\n10,1.2"))
	r := NewReader(f)
	r.Loop(func(e struct {
		Int   int     `name:"a"`
		Float float32 `name:"c"`
	}) error {
		t.Error("The callback of Look is invoked unexpectedly")
		return nil
	})
	err := r.Done()
	if err == nil || err.Error() != "c did not appear in the first line" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLoopWithSlice(t *testing.T) {
	f := bytes.NewReader([]byte("10,20\n30,40"))
	r := NewReader(f)
	var rows [][]int
	r.Loop(func(row []int) error {
		rows = append(rows, row)
		return nil
	})
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expected := [][]int{{10, 20}, {30, 40}}
	if !reflect.DeepEqual(rows, expected) {
		t.Errorf("Expected %#v but got %#v", expected, rows)
	}
}

func TestLoopBreak(t *testing.T) {
	f := bytes.NewReader([]byte("10,20\n30,40"))
	r := NewReader(f)
	var rows [][]int
	r.Loop(func(row []int) error {
		rows = append(rows, row)
		return Break
	})
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expected := [][]int{{10, 20}}
	if !reflect.DeepEqual(rows, expected) {
		t.Errorf("Expected %#v but got %#v", expected, rows)
	}
}

func TestLoopBreakWithError(t *testing.T) {
	f := bytes.NewReader([]byte("10,20\n30,40"))
	r := NewReader(f)
	e := errors.New("error")
	var rows [][]int
	r.Loop(func(row []int) error {
		rows = append(rows, row)
		return e
	})
	if err := r.Done(); err != e {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	expected := [][]int{{10, 20}}
	if !reflect.DeepEqual(rows, expected) {
		t.Errorf("Expected %#v but got %#v", expected, rows)
	}
}

func TestRead(t *testing.T) {
	f := bytes.NewReader([]byte("10,1.2\n20,2.3\n30,3.4"))
	r := NewReader(f)
	var ints []int
	var floats []float32
	var e struct {
		Int   int     `index:"0"`
		Float float32 `index:"1"`
	}
	for r.Read(&e) {
		ints = append(ints, e.Int)
		floats = append(floats, e.Float)
	}
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expectedInt := []int{10, 20, 30}
	expectedFloat := []float32{1.2, 2.3, 3.4}
	if !reflect.DeepEqual(expectedInt, ints) {
		t.Errorf("Unexpected %#v but got %#v", expectedInt, ints)
	}
	if !reflect.DeepEqual(expectedFloat, floats) {
		t.Errorf("Unexpected %#v but got %#v", expectedFloat, floats)
	}
}

func TestReadWithName(t *testing.T) {
	f := bytes.NewReader([]byte("a,b\n10,1.2\n20,2.3\n30,3.4"))
	r := NewReader(f)
	var ints []int
	var floats []float32
	var e struct {
		Int   int     `name:"a"`
		Float float32 `name:"b"`
	}
	for r.Read(&e) {
		ints = append(ints, e.Int)
		floats = append(floats, e.Float)
	}
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expectedInt := []int{10, 20, 30}
	expectedFloat := []float32{1.2, 2.3, 3.4}
	if !reflect.DeepEqual(expectedInt, ints) {
		t.Errorf("Unexpected %#v but got %#v", expectedInt, ints)
	}
	if !reflect.DeepEqual(expectedFloat, floats) {
		t.Errorf("Unexpected %#v but got %#v", expectedFloat, floats)
	}
}

func TestReadIndexOutOfRange(t *testing.T) {
	f := bytes.NewReader([]byte("10,1.2\n20,2.3"))
	r := NewReader(f)
	var e struct {
		Int   int     `index:"0"`
		Float float32 `index:"2"`
	}
	for r.Read(&e) {
		t.Errorf("r.Read returned true unexpectedly with %#v", e)
	}
	err := r.Done()
	if err == nil || err.Error() != "Accessed index 2 though the size of the row is 2" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReadMissingColumn(t *testing.T) {
	f := bytes.NewReader([]byte("a,c\n10,1.2"))
	r := NewReader(f)
	var e struct {
		Int   int     `name:"a"`
		Float float32 `name:"b"`
	}
	for r.Read(&e) {
		t.Errorf("r.Read returned true unexpectedly with %#v", e)
	}
	err := r.Done()
	if err == nil || err.Error() != "b did not appear in the first line" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReadWithSlice(t *testing.T) {
	f := bytes.NewReader([]byte("10,20\n30,40"))
	r := NewReader(f)
	var rows [][]int
	var row []int
	for r.Read(&row) {
		rows = append(rows, row)
	}
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expected := [][]int{{10, 20}, {30, 40}}
	if !reflect.DeepEqual(rows, expected) {
		t.Errorf("Expected %#v but got %#v", expected, rows)
	}
}

func TestReadAllStruct(t *testing.T) {
	f := bytes.NewReader([]byte("10,2.3\n30,4.5"))
	r := NewReader(f)
	type entry struct {
		Int   int     `index:"0"`
		Float float32 `index:"1"`
	}
	var s []entry
	r.ReadAll(&s)
	expected := []entry{{Int: 10, Float: 2.3}, {Int: 30, Float: 4.5}}
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(expected, s) {
		t.Errorf("Expected %v but got %v", expected, s)
	}
}

func TestReadAllSlice(t *testing.T) {
	f := bytes.NewReader([]byte("10,20\n30,40"))
	r := NewReader(f)
	var s [][]int
	r.ReadAll(&s)
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expected := [][]int{{10, 20}, {30, 40}}
	if !reflect.DeepEqual(expected, s) {
		t.Errorf("Expected %v but got %v", expected, s)
	}
}

func TestEncTag(t *testing.T) {
	f := bytes.NewReader([]byte("10,10,010\n20,20,020"))
	r := NewReader(f)
	var ints0 []int
	var ints1 []int
	var ints2 []int
	var e struct {
		Int0 int `index:"0" enc:"hex"`
		Int1 int `index:"1" enc:"oct"`
		Int2 int `index:"2" enc:"deci"`
	}
	for r.Read(&e) {
		ints0 = append(ints0, e.Int0)
		ints1 = append(ints1, e.Int1)
		ints2 = append(ints2, e.Int2)
	}
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expectedInt0 := []int{16, 32}
	expectedInt1 := []int{8, 16}
	expectedInt2 := []int{10, 20}
	if !reflect.DeepEqual(expectedInt0, ints0) {
		t.Errorf("Unexpected %#v but got %#v", expectedInt0, ints0)
	}
	if !reflect.DeepEqual(expectedInt1, ints1) {
		t.Errorf("Unexpected %#v but got %#v", expectedInt1, ints1)
	}
	if !reflect.DeepEqual(expectedInt2, ints2) {
		t.Errorf("Unexpected %#v but got %#v", expectedInt2, ints2)
	}
}

func TestNewDecoder(t *testing.T) {
	d, err := newDecoder(reflect.TypeOf(struct {
		Name int `name:"name"`
		Age  int `name:"age"`
	}{}))
	if err != nil {
		t.Error(err)
	}
	if !d.needHeader() {
		t.Error("Unexpected")
	}
}

func TestNewDecoder_IndexAndName(t *testing.T) {
	_, err := newDecoder(reflect.TypeOf(struct {
		Name int `name:"name"`
		Age  int `index:"0"`
	}{}))
	if err == nil || err.Error() != "Fields with name and fields with index are mixed" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewDecoder_NoStructTag(t *testing.T) {
	_, err := newDecoder(reflect.TypeOf(struct {
		Name int
		Age  int
	}{}))
	if err == nil || err.Error() != "Please specify name or index to the struct field: Name\nPlease specify name or index to the struct field: Age" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewDecoder_InvalidIndex(t *testing.T) {
	_, err := newDecoder(reflect.TypeOf(struct {
		Name int `index:"-1"`
		Age  int `index:"hello"`
	}{}))
	if err == nil || err.Error() != "Failed to parse index of field Name: \"-1\"\nFailed to parse index of field Age: \"hello\"" {
		t.Errorf("Unexpected error: %q", err)
	}
}

func TestLineNumber(t *testing.T) {
	f := bytes.NewReader([]byte("10,1.2\n20,2.3\n30,3.4"))
	r := NewReader(f)
	var ints []int
	var lineno []int
	r.Loop(func(e struct {
		Int   int     `index:"0"`
	}) error {
		ints = append(ints, e.Int)
		lineno = append(lineno, r.LineNumber())
		return nil
	})
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	expectedInt := []int{10, 20, 30}
	expectedLineno := []int{1, 2, 3}
	if !reflect.DeepEqual(expectedInt, ints) {
		t.Errorf("Expected %#v but got %#v", expectedInt, ints)
	}
	if !reflect.DeepEqual(expectedLineno, lineno) {
		t.Errorf("Expected %#v but got %#v", expectedLineno, lineno)
	}
}
