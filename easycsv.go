package easycsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Break is the error returned by the callback passed to Loop to terminate the loop.
var Break = errors.New("break")

// Reader provides a convenient interface for reading csv.
type Reader struct {
	// csv.Reader. To read content from csv, use readLine.
	csv    *csv.Reader
	closer io.Closer
	done   bool
	// An error occurred while processing csv. io.EOF is stored when csv is reached to the end.
	err error
	opt Option

	// Used from readLine.
	lineno    int
	firstLine []string
	cur       []string
}

func newCSVReader(r io.Reader, opt Option) *csv.Reader {
	cr := csv.NewReader(r)
	if opt.Comma != 0 {
		cr.Comma = opt.Comma
	}
	if opt.Comment != 0 {
		cr.Comment = opt.Comment
	}
	if opt.LazyQuotes {
		cr.LazyQuotes = opt.LazyQuotes
	}

	return cr
}

// NewReader returns a new Reader to read CSV from r.
func NewReader(r io.Reader, opts ...Option) *Reader {
	opt, err := mergeOptions(opts)
	if err != nil {
		return &Reader{err: err}
	}
	rd := Reader{
		csv: newCSVReader(r, opt),
		opt: opt,
	}
	return &rd
}

// NewReadCloser returns a new Reader to read CSV from r.
// Reader instantiated with NewReadCloser closes r automatically when Done() is called.
func NewReadCloser(r io.ReadCloser, opts ...Option) *Reader {
	opt, err := mergeOptions(opts)
	if err != nil {
		return &Reader{err: err}
	}
	return &Reader{
		csv:    newCSVReader(r, opt),
		opt:    opt,
		closer: r,
	}
}

// NewReaderFile returns a new Reader to read CSV from the file path.
func NewReaderFile(path string, opts ...Option) *Reader {
	f, err := os.Open(path)
	if err == nil {
		return NewReadCloser(f, opts...)
	}
	return &Reader{err: err}
}

// readLine reads a line from r.csv and update r.err, r.cur, r.lineno and r.firstLine.
// readLine does not update r.err. io.EOF is returned when csv reached to the end.
func (r *Reader) readLine() {
	line, err := r.csv.Read()
	if err != nil {
		r.err = err
		return
	}
	r.cur = line
	r.lineno++
	if r.lineno == 1 {
		r.firstLine = line
	}
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Loop reads from r until an error or EOF and invokes body everytime it reads a line.
// body should be a function which has no return value or returns bool or error.
// If body returns false or an error, Loop stops reading r.
// If body does not have a return value, Loop does not stop until an error or EOF.
// If body returns error and you want to terminate Loop without reporting an error, returns easycsv.Break in body.
//
// Also, body must receive one argument. The argument must be a pointer to a struct, a struct or a pointer to a slice.
// The line of csv is automatically converted to the struct or the slice based on the rule described above.
//
// Loop returns an error if it encounters an error and exits the loop.
func (r *Reader) Loop(body interface{}) (err error) {
	defer func() { err = r.Done() }()
	if r.err != nil {
		return
	}
	if body == nil {
		r.err = errors.New("The argument of Loop must not be nil.")
		return
	}
	v := reflect.TypeOf(body)
	if v.Kind() != reflect.Func {
		r.err = fmt.Errorf("The argument of Loop must be func but got %v", v.Kind())
		return
	}
	if v.NumIn() != 1 || v.NumOut() > 1 {
		r.err = fmt.Errorf("The function passed to Loop must receive one argument and return one or zero value")
		return
	}
	if v.NumOut() > 0 {
		if out := v.Out(0); out.Kind() != reflect.Bool && out != errorType {
			r.err = fmt.Errorf("The function passed to Loop must return error or bool")
			return
		}
	}
	in := v.In(0)
	var inStruct reflect.Type
	if in.Kind() == reflect.Struct {
		inStruct = in
	} else if in.Kind() == reflect.Ptr && in.Elem().Kind() == reflect.Struct {
		inStruct = in.Elem()
	} else if in.Kind() == reflect.Slice {
		inStruct = in
	} else {
		r.err = fmt.Errorf("The function passed to Loop must receive a struct, a pointer to a struct or a slice")
		return
	}
	if in.Kind() != reflect.Slice {
		numf := inStruct.NumField()
		if numf == 0 {
			r.err = errors.New("The struct passed to Loop must have at least one field")
			return
		}
	}
	dec, err := newDecoder(r.opt, inStruct)
	if err != nil {
		r.err = err
		return
	}
	if dec.needHeader() {
		if r.lineno == 0 {
			// Loop quits immediately if the csv is empty.
			r.readLine()
			if r.err != nil {
				return
			}
		}
		err = dec.consumeHeader(r.firstLine)
		if err != nil {
			r.err = err
			return
		}
	}
	for {
		r.readLine()
		if r.err != nil {
			break
		}
		p := reflect.New(inStruct)
		if err := dec.decode(r.cur, p); err != nil {
			r.err = err
			break
		}
		arg := p
		if in.Kind() == reflect.Struct || in.Kind() == reflect.Slice {
			arg = p.Elem()
		}
		rets := reflect.ValueOf(body).Call([]reflect.Value{arg})
		if len(rets) == 0 {
			continue
		}
		ret := rets[0]
		if ret.Kind() == reflect.Bool {
			if ret.Bool() {
				continue
			} else {
				break
			}
		}
		if ret.IsNil() {
			// body returned nil error.
			continue
		}
		err = rets[0].Interface().(error)
		if err == nil {
			panic("err must not be nil if I understand reflect spec correctly")
		}
		if err != Break {
			r.err = err
		}
		break
	}
	return
}

// Read reads one line from csv and store values in the line to e.
// The argument e must be a pointer to a struct or a pointer to a slice.
// Read returns false when it encounters an error or EOF.
func (r *Reader) Read(e interface{}) bool {
	if r.err != nil {
		return false
	}
	if e == nil {
		r.err = errors.New("The argument of Loop must not be nil.")
		return false
	}
	t := reflect.TypeOf(e)
	if t.Kind() != reflect.Ptr {
		r.err = fmt.Errorf("The argument of Read must be a pointer to a struct or a pointer to a slice, but got %v", t.Kind())
		return false
	}
	if t.Elem().Kind() != reflect.Struct && t.Elem().Kind() != reflect.Slice {
		r.err = fmt.Errorf("The argument of Read must be a pointer to a struct or a pointer to a slice, but got a pointer to %v", t.Elem().Kind())
		return false
	}
	decoder, err := newDecoder(r.opt, t.Elem())
	if err != nil {
		r.err = err
		return false
	}
	if decoder.needHeader() {
		if r.lineno == 0 {
			// Loop quits immediately if the csv is empty.
			r.readLine()
			if r.err != nil {
				return false
			}
		}
		if err := decoder.consumeHeader(r.firstLine); err != nil {
			r.err = err
			return false
		}
	}
	r.readLine()
	if r.err != nil {
		return false
	}
	// TODO: Append the line number to the error message.
	r.err = decoder.decode(r.cur, reflect.ValueOf(e))
	return r.err == nil
}

// ReadAll reads all rows from csv and store it into the slice s.
// s must be a pointer to a slice of a struct (e.g. *[]entry) or a pointer to a slice of primitive types (e.g. *[][]int).
// ReadAll reports an error if it encounters an error while reading the input.
// Also, ReadAll closes the file behind r automatically.
func (r *Reader) ReadAll(s interface{}) (err error) {
	defer func() { err = r.Done() }()
	// TODO: Consolidate code with Read.
	if s == nil {
		r.err = errors.New("The argument of ReadAll must not be nil.")
		return
	}
	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		r.err = fmt.Errorf("The argument of ReadAll must be a pointer to a slice of a slice or a pointer to a slice of a struct, but got %v", t)
		return
	}
	et := t.Elem().Elem()
	if et.Kind() != reflect.Struct && et.Kind() != reflect.Slice {
		r.err = fmt.Errorf("The argument of ReadAll must be a pointer to a slice of a slice or a pointer to a slice of a struct, but got %v", t)
		return
	}
	decoder, err := newDecoder(r.opt, et)
	if err != nil {
		r.err = err
		return
	}
	if decoder.needHeader() {
		if r.lineno == 0 {
			// Loop quits immediately if the csv is empty.
			r.readLine()
			if r.err != nil {
				return
			}
		}
		decoder.consumeHeader(r.firstLine)
	}
	for {
		r.readLine()
		if r.err != nil {
			return
		}
		p := reflect.New(et)
		v := reflect.ValueOf(s).Elem()
		err := decoder.decode(r.cur, p)
		v.Set(reflect.Append(v, p.Elem()))
		if err != nil {
			r.err = err
			break
		}
		// TODO: Append the line number to the error message.
	}
	return
}

func (r *Reader) nonEOFError() error {
	if r.err == nil || r.err == io.EOF {
		return nil
	}
	return r.err
}

// Done returns the first non-EOF error that was encountered by the Reader.
// Done also closes the internal Closer if the Reader is instantiated with NewReaderCloser.
//
// You need to use Done when you read CSV with Read method to check errors and close files behind.
// You don't use Done when you read CSV with ReadAll and Loop because they call Done internally.
func (r *Reader) Done() error {
	if r.done {
		return r.nonEOFError()
	}
	r.done = true
	if r.closer != nil {
		if cerr := r.closer.Close(); r.err == nil || r.err == io.EOF {
			r.err = cerr
		}
	}
	return r.nonEOFError()
}

// DoneDefer does the same thing as Done does. But it outputs an error to the argument.
// DoneDefer does not overwrite an error if an error is already stored in err.
// DoneDefer is useful to call Done from defer statement.
func (r *Reader) DoneDefer(err *error) {
	e := r.Done()
	if *err == nil && e != nil {
		*err = e
	}
}

// LineNumber returns the current position of r in the input (1-based line number).
// The behavior of LineNumber is undefined if it is called before Read operations or
// after r reached to EOF or an error.
func (r *Reader) LineNumber() int {
	return r.lineno
}

type rowDecoder interface {
	decode(s []string, out reflect.Value) error
	needHeader() bool
	consumeHeader([]string) error
}

func validateCustomConverter(conv interface{}, enc string, field reflect.StructField, errs *[]string) bool {
	convType := reflect.TypeOf(conv)
	if convType.Kind() != reflect.Func {
		*errs = append(*errs, fmt.Sprintf("The custom decoder for Encoding %q must be a function", enc))
		return false
	}
	ok := true
	if numin := convType.NumIn(); numin != 1 {
		*errs = append(*errs, fmt.Sprintf("The custom decoder for Encoding %q must receive an arg, but receives %d args", enc, numin))
		ok = false
	} else if convType.In(0).Kind() != reflect.String {
		*errs = append(*errs, fmt.Sprintf("The custom decoder for Encoding %q must receive a string, but receives %v", enc, convType.In(0)))
		ok = false
	}
	// TODO: Supports custom decoders that does not return an error.
	if numout := convType.NumOut(); numout != 2 {
		*errs = append(*errs, fmt.Sprintf("The custom decoder for Encoding %q must return two values, but returns %d values", enc, numout))
		ok = false
	} else {
		if convType.Out(0) != field.Type {
			*errs = append(*errs, fmt.Sprintf("The type of field %q is %v, but enc %q returns %q", field.Name, field.Type, enc, convType.Out(0)))
			ok = false
		}
		if convType.Out(1) != errorType {
			*errs = append(*errs, fmt.Sprintf("The second return value of the custom decoder for %q must be error", enc))
			ok = false
		}
	}
	return ok
}

func parseStructTag(
	opt Option,
	field reflect.StructField,
	fieldIdx int,
	nameMap map[string]int,
	idxMap map[int]int,
	converters *[]reflect.Value,
	errors *[]string) {
	tag := field.Tag
	name := tag.Get("name")
	index := tag.Get("index")
	if name == "" && index == "" {
		*errors = append(*errors, fmt.Sprintf("Please specify name or index to the struct field: %s", field.Name))
		return
	}
	if name != "" && index != "" {
		*errors = append(*errors, fmt.Sprintf("Please specify name or index to the struct field: %s", field.Name))
		return
	}
	var conv interface{}
	enc := tag.Get("enc")
	if enc != "" {
		if opt.Decoders != nil && opt.Decoders[enc] != nil {
			conv = opt.Decoders[enc]
			if !validateCustomConverter(conv, enc, field, errors) {
				conv = nil
			}
		} else {
			pre := predefinedDecoders[enc]
			// TODO: Test these errors.
			if pre != nil {
				conv = pre(field.Type)
				if conv == nil {
					*errors = append(*errors, fmt.Sprintf("Encoding %q does not support %v", enc, field.Type))
				}
			} else {
				*errors = append(*errors, fmt.Sprintf("Encoding %q is not defined", enc))
				return
			}
		}
	}
	if conv == nil {
		var err error
		conv, err = createConverterFromType(opt, field.Type)
		if err != nil {
			*errors = append(*errors, err.Error())
		}
	}
	if conv == nil {
		*errors = append(*errors, fmt.Sprintf("Unexpected field type for %s: %s", field.Name, field.Type))
		return
	}
	*converters = append(*converters, reflect.ValueOf(conv))
	if name != "" {
		nameMap[name] = fieldIdx
		return
	}
	i, err := strconv.Atoi(index)
	if err != nil || i < 0 {
		*errors = append(*errors, fmt.Sprintf("Failed to parse index of field %s: %q", field.Name, index))
		return
	}
	idxMap[i] = fieldIdx
}

func newDecoder(opt Option, t reflect.Type) (rowDecoder, error) {
	if t.Kind() == reflect.Struct {
		return newStructDecoder(opt, t)
	} else if t.Kind() == reflect.Slice {
		return newSliceDecoder(opt, t)
	}
	panic("newDecoder must be called with struct or slice.")
}

func newSliceDecoder(opt Option, t reflect.Type) (rowDecoder, error) {
	elem := t.Elem()
	c, err := createConverterFromType(opt, elem)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, fmt.Errorf("Failed to create a converter for %v", t)
	}
	return &sliceRowDecoder{
		elemType:  elem,
		converter: reflect.ValueOf(c),
	}, nil
}

type sliceRowDecoder struct {
	elemType  reflect.Type
	converter reflect.Value
}

func (d *sliceRowDecoder) needHeader() bool             { return false }
func (d *sliceRowDecoder) consumeHeader([]string) error { return nil }
func (d *sliceRowDecoder) decode(s []string, out reflect.Value) error {
	slicePtr := reflect.New(reflect.SliceOf(d.elemType))
	for _, e := range s {
		rets := d.converter.Call([]reflect.Value{reflect.ValueOf(e)})
		if len(rets) != 2 {
			panic("converter must return two values.")
		}
		if !rets[1].IsNil() {
			return rets[1].Interface().(error)
		}
		slicePtr.Elem().Set(reflect.Append(slicePtr.Elem(), rets[0]))
	}
	out.Elem().Set(slicePtr.Elem())
	return nil
}

func newStructDecoder(opt Option, t reflect.Type) (rowDecoder, error) {
	if t.NumField() == 0 {
		return nil, errors.New("The struct has no field")
	}
	v := reflect.New(t).Elem()
	var unexported []string
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			unexported = append(unexported, t.Field(i).Name)
		}
	}
	if unexported != nil {
		return nil, fmt.Errorf("The struct passed to Loop must not have unexported fields: %s", strings.Join(unexported, ", "))
	}

	var tagErrors []string
	nameMap := make(map[string]int)
	idxMap := make(map[int]int)
	var converters []reflect.Value
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		parseStructTag(opt, f, i, nameMap, idxMap, &converters, &tagErrors)
	}
	if len(nameMap) != 0 && len(idxMap) != 0 {
		tagErrors = append(tagErrors, "Fields with name and fields with index are mixed")
	}
	if tagErrors != nil {
		return nil, errors.New(strings.Join(tagErrors, "\n"))
	}
	if len(converters) != t.NumField() {
		panic("converters size mismatch")
	}
	if len(nameMap) != 0 {
		idxMap = nil
	} else {
		nameMap = nil
	}
	return &structRowDecoder{
		structType: t,
		converters: converters,
		names:      nameMap,
		indice:     idxMap,
	}, nil
}

type structRowDecoder struct {
	structType reflect.Type
	converters []reflect.Value
	names      map[string]int
	indice     map[int]int
}

func (d *structRowDecoder) consumeHeader(header []string) error {
	indice := make(map[int]int)
	for i, col := range header {
		idx, ok := d.names[col]
		if !ok {
			continue
		}
		indice[i] = idx
		delete(d.names, col)
	}
	d.indice = indice
	if len(d.names) != 0 {
		var unused []string
		for n := range d.names {
			unused = append(unused, n)
		}
		return fmt.Errorf("%s did not appear in the first line", strings.Join(unused, ", "))
	}
	d.names = nil
	return nil
}

func (d *structRowDecoder) decode(row []string, out reflect.Value) error {
	// TODO: Reset with zero first.
	for i, j := range d.indice {
		if i >= len(row) {
			return fmt.Errorf("Accessed index %d though the size of the row is %d", i, len(row))
		}
		rets := d.converters[j].Call([]reflect.Value{reflect.ValueOf(row[i])})
		if len(rets) != 2 {
			panic("converter must return two values.")
		}
		if !rets[1].IsNil() {
			return rets[1].Interface().(error)
		}
		out.Elem().Field(j).Set(rets[0])
	}
	return nil
}

func (d *structRowDecoder) needHeader() bool {
	return d.names != nil
}
