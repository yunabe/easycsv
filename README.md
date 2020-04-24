# easycsv [![Build Status](https://travis-ci.org/yunabe/easycsv.svg?branch=master)](https://travis-ci.org/yunabe/easycsv) [![Go Report Card](https://goreportcard.com/badge/github.com/yunabe/easycsv)](https://goreportcard.com/report/github.com/yunabe/easycsv) [![Binder](https://mybinder.org/badge_logo.svg)](https://mybinder.org/v2/gh/yunabe/easycsv-binder/master?filepath=example.ipynb)
easycsv package provides API to read CSV files in Go (golang) easily.

# Installation
```
go get -u github.com/yunabe/easycsv
```

# Features
- You can read CSV files with less boilerplate code because `easycsv` provides a consice error API.
- `easycsv` automatically converts CSV rows into your custom structs.
- Of course, you can handle TSV and other CSV-like formats by customizing `easycsv.Reader`.

# Links
- [yunabe/easycsv godoc](https://godoc.org/github.com/yunabe/easycsv)
- [yunabe/easycsv GitHub](https://github.com/yunabe/easycsv)

# Quick Tour

## Read a CSV file to a struct
```golang
r := easycsv.NewReaderFile("testdata/sample.csv")
var entry struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}
for r.Read(&entry) {
	fmt.Print(entry)
}
if err := r.Done(); err != nil {
	log.Fatalf("Failed to read a CSV file: %v", err)
}
```

## Read a CSV file with Loop
```golang
r := easycsv.NewReaderFile("testdata/sample.csv")
err := r.Loop(func(entry *struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}) error {
	fmt.Print(entry)
	return nil
})
if err != nil {
	log.Fatalf("Failed to read a CSV file: %v", err)
}
```

# Usages

## NewReader
The core component of easycsv is [`Reader`](https://godoc.org/github.com/yunabe/easycsv#Reader).
You can create a new `Reader` instance from `io.Reader`, `io.ReadCloser` and a file path.

- [NewReader](https://godoc.org/github.com/yunabe/easycsv#NewReader)
  - Create `easycsv.Reader` from `io.Reader`.
- [NewReadCloser](https://godoc.org/github.com/yunabe/easycsv#NewReadCloser)
  - Create `easycsv.Reader` from `io.ReadCloser`.
- [NewReaderFile](https://godoc.org/github.com/yunabe/easycsv#NewReaderFile)
  - Create `easycsv.Reader` from a file path.

The Reader created by NewReadCloser and NewReaderFile closes the file automatically when the Reader is finished.
So you do not need to close files manually and you can omit error handling code for closing files.

## Read
There are three methods to read CSV with `easycsv.Reader`. Read, Loop and ReadAll.
We are looking into [`Read`](https://godoc.org/github.com/yunabe/easycsv#Reader.Read) method first, which is the most basic and naive way to read CSV with Reader.

```golang
func (r *Reader) Read(e interface{}) bool
```

[`Read`](https://godoc.org/github.com/yunabe/easycsv#Reader.Read) receives a pointer to a struct (e.g. `*myStruct`) or a pointer to a slice of a primitive type (e.g. `*[]int`).
If it reads a new row from CSV successufly, it stores the row into `e` and returns `true`.
If `Reader` reaches to `EOF` or it fails to read a new row for some reasons, it returns `false`.
`Read` returns `false` for various reasons. To check the reason, you have to call `Done()` subsequently.
`Done` returns an error if `Read` encountered an error.
`Done` returns `nil` if `Read` returned `false` because it reached to `EOF`.

You can pass two types of pointers to Read. A pointer of a struct (e.g. `*myStruct`) or  a pointer of a slice of primitive typs (e.g. `*[]int`). Passing a pointer of a struct is more convenient.
When you use a struct, you need to specify how to map CSV columns to the struct's field using struct field's tags.
Here are examples:

```golang
var entry struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}

var entry struct {
	Name string `name:"name"`
	Age  int    `name:"age"`
}
```

You can use `index` tag or `name` tag to specify the mapping.
When `index` is used, Read maps `index`-th (0-based) column to the field.
In the first example, the frist column is mapped to Name and the second column is mapped to Age.
When `name` is used, Read uses the first line of CSV as a header with column names and maps columns to fields based on the column names in the header. In the second example, give that the content of CSV is the following,

```csv
age,name
10,Alice
20,Bob
```

the frist column is mapped to Age and the second column is mapped to Name. So `{Alice 10}` and `{Bob 20}` are stored to the struct respectively. You can not use both `index` tag and `name` tag in the same struct. Read reports an error in that case.

If you pass a pointer to a slice to Read, Read converts CSV row into the slice and fills it to the argument.
If the argument `e` is invalid, Read returns false immediately and the reason of the error is reported by `Done()`.

The conversion from CSV row (string) to the given field type (int, float32, bool, etc...) is handled in Reader automatically. If you want to customize the conversion, see [Custom encoding](#custom-encoding) section below.

When you read CSV with `Read` methods, you have to always call `Done()` subsequently to (1) check the error and (2) close the file behind the Reader. If you forget to call `Done()`, the error will be completely gone.
Do not forget to call `Done` after `Read`.

## Loop
```golang
func (r *Reader) Loop(body interface{}) (err error)
```

[Loop](https://godoc.org/github.com/yunabe/easycsv#Loop) reads CSV line by line and executes `body` with a line everytime it reads a line.
`body` must be a function that receives a struct (e.g. `myStruct`), a pointer of a struct (e.g. `*myStruct`) or a slice of primitives (e.g. `[]int`).
A line of CSV is automatically converted to the argument of `body` when Loop reads the line and passed to `body`.
Also, `body` must be a function that returns `bool`, `error` or no return value.
If `body` is a function that returns `bool`, Loop stops reading CSV at the line where `body` returns false.
If `body` is a function that returns `error`, Loop stops reading CSV when `body` retruns an error.
Loop does not stop until it reached to the end if `body` has no return value.
If `body` retuns an error, Loop quits and reports the error.

When `Loop` ends, it invokes `Done` and closes internal files automatically.
So, you do not need to call Done after Loop.
Loop returns the first error if it encounters errors. It returns `nil` if everything goes well.
Do not forget to handle the error returned by Loop.

The example below shows how to use Loop with a function which returns `error`.
This code reads CSV until Loop reaches to EOF or an entry with Age < 0 is found in the CSV.

```golang
err := r.Loop(func(entry *struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}) error {
	fmt.Println(entry)
	if Age < 0 {
		return errors.New("Age mustn't be negative")
	}
})
if err != nil {
	log.Fatalf("Failed to read a CSV file: %v", err)
}
```

## ReadAll
```golang
func (r *Reader) ReadAll(s interface{}) (err error)
```

[ReadAll](https://godoc.org/github.com/yunabe/easycsv#ReadAll) reads a CSV input to the end and convert all rows into the slice passed as an argument.
The argument `s` must be a pointer of a slice of a struct (`*[]myStruct`) or a pointer of a slice of a slice (`*[][]int`).
Aside from that, the same rules of Read are applied to ReadAll. You need to specify how to map columns to struct fields using struct field's tag.

Like `Loop`, you do not need to call `Done` after ReadAll. ReadAll returns the first error if it encounters errors. It returns `nil` if everything goes well.
Do not forget to handle the error returned by ReadAll.

```golang
var entry []struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}
err := r.ReadAll(&entry);
```

# Option
To control the behavior of Reader, you can pass Option to NewReader methods.

NewReader methods receive Option as a variadic parameter `opts`. `opts` is a variadic parameter so that we can omit `opts` from parameters when we call NewReader methods without changing Option.
Thus, you don't need to pass multiple Option to NewReader methods although you can pass as many Option as you want.

## Comma
Like [csv.Reader](https://golang.org/pkg/encoding/csv/#Reader) in the standard library, you can change the deliminator of CSV by specifying `Comma` option. For example, if you set `'\t'` to Comma, Reader reads a file as a TSV file.

## Comment
Comment, if not 0, is the comment character. Lines beginning with the character without preceding whitespace are ignored.

## FieldsPerRecord
In the standard library [csv.Reader](https://golang.org/pkg/encoding/csv/#Reader), an option `FieldsPerRecord` is available to define the number of fields allowed per CSV record. If you set a value that is not 0 to `FieldsPerRecord`, this option will be updated.

# Customizing decoders
By default, easycsv converts strings in CSV to integers, floats and bool automatically based on the types of struct fields and slices.

- Integers are parsed with `strconv.ParseInt` and unsigned integers are parsed with `strconv.ParseUint`.
  easycsv parses inputs as decimals by default. But it parses inputs as hex if inputs have `"0x"` prefix and
  as octal if inputs have `"0"` prefix (`"0xff"` → 255, `"077"` → 63).
- Floats are parsed with `strconv.ParseFloat`.
- bool is parsed with `strconv.ParseBool`.

You can customize how to decode strings in CSV to values by specifying `enc` attribute to struct fields.

## Predefined encoding
easycsv has three predefined custom encoding for integers.

- `deci` - Parses inputs as decimal integers even if the inputs are prefixed with `"0"`.
- `hex` - Parses inputs as hex integers.
- `oct`- Parses inputs as oct integers.

## Custom encoding
Also, you can use custom encodings in easycsv.

To use custom encodings:
- Define a func that convert strings to your custom types. This func must receive a string and returns (custom-type, error).
- Register the func to Option.Decoders.
- Specify the registered func name with `enc` struct-field attribute.

```golang
r := NewReader(bytes.NewBufferString("name,birthday\nAlice,1980-12-30\nBob,1975-06-09"),
	Option{
		Decoders: map[string]interface{}{
			"date": func(s string) (time.Time, error) {
				return time.Parse("2006-01-02", s)
			},
	},
	})
var entry struct {
	Name  string    `name:"name"`
	Birth time.Time `name:"birthday" enc:"date"`
}
for r.Read(&entry) {
	fmt.Print(entry)
}
if err := r.Done(); err != nil {
	fmt.Printf("Failed: %v\n", err)
}
// Output: {Alice 1980-12-30 00:00:00 +0000 UTC}{Bob 1975-06-09 00:00:00 +0000 UTC}
```

## Customizing decoders for types
You can also define how to convert strings into specific types in easycsv by using Option.TypeDecoders option. Option.TypeDecoders is similar to Option.Decoders. The key is `reflect.Type` and the value is a function to convert strings to the specific type.
Reader uses the functions registered to Option.TypeDecoders instead of default converters when it converts rows in CSV into those types.

The following example shows how to define a converter for `time.Time` with Option.TypeDecoders.

```golang
r := NewReader(bytes.NewReader([]byte("2017-01-02,2016-02-03\n2015-03-04,2014-04-05")),
	Option{TypeDecoders: map[reflect.Type]interface{}{
		reflect.TypeOf(time.Time{}): func(s string) (time.Time, error) {
			return time.Parse("2006-01-02", s)
		},
	}})
var entry []time.Time
for r.Read(&entry) {
	for _, e := range entry {
		fmt.Print(e.Format("2006/1/2"), ";")
	}
}
if err := r.Done(); err != nil {
	fmt.Print(err)
}
```
