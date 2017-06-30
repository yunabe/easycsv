# easycsv
easycsv package provides API to read CSV file in Go (golang).

# Installation
```
go get -u github.com/yunabe/easycsv
```

# Features
- You can read CSV files with less boilerplate code because `easycsv` provides a consice error API.
- `easycsv` automatically converts CSV rows into your custom structs.
- Of course, you can handle TSV and other CSV-like formats by customizing `easycsv.Reader`.

# Quick Look

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

## Read a CSV with Loop
```golang
r := easycsv.NewReaderFile("testdata/sample.csv")
r.Loop(func(entry *struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}) error {
	fmt.Print(entry)
	return nil
})
if err := r.Done(); err != nil {
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
So you do not need to close files manually and you are able to omit an error handling code for closing files.

## Read
There are three methods to read CSV with `easycsv.Reader`. Read, Loop and ReadAll.
We are looking into [`Read`](https://godoc.org/github.com/yunabe/easycsv#Reader.Read) method first, which is the most basic and naive way to read CSV with Reader.

```golang
func (r *Reader) Read(e interface{}) bool
```

[`Read`](https://godoc.org/github.com/yunabe/easycsv#Reader.Read) receives a pointer to a struct (e.g. `*myStruct`) or a pointer to a slice of a primitive type (e.g. `*[]int`).
If it reads a new row from CSV successufly, it stores the row into `e` and returns `true`.
If `Reader` reached to `EOF` or it fails to read a new row for some reasons, it returns `false`.
`Read` returns `false` for various reasons. To know the reason, you have to call `Done()` subsequently.
`Done` returns an error if `Read` encountered an error.
`Done` returns `nil` if `Read` returned `false` because it reached to `EOF`.

You can pass two types of pointers to Read. A pointer to a struct (e.g. `*myStruct`) or  a pointer to a slice of primitive typs (e.g. `*[]int`). Passing a pointer to a struct is more convenient.
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

The conversion from CSV row (string) to the given field type (int, float32, bool, etc...) is handled in Reader automatically.

When you read CSV with `Read` methods, you have to always call `Done()` subsequently to (1) check the error and (2) close the file behind the Reader when it is instantiated with `NewReadCloser` or `NewReaderFile`. If you forget to call `Done()`, the error will be completely gone.

## Loop
```golang
func (r *Reader) Loop(body interface{})
```

Loop reads CSV a line by line and executes `body` with a line everytime it reads a line.
`body` must be a function that receives a struct (e.g. `myStruct`), a pointer of a struct (e.g. `*myStruct`) or a slice of primitives (e.g. `[]int`).
The line of CSV is automatically converted to the argument of `body` when Loop reads the line before it calls `body`.

Also, `body` must be a function that returns `bool`, `error` or no return value.
If `body` is a function that returns `bool`, Loop stops reading CSV at the line where `body` returns false.
If `body` is a function that returns `error`, Loop stops reading CSV when `body` retruns an error.
Loop does not stop until it reached to the end if `body` has no return value.
If `body` retuns an error, Loop quits and the error is reported when `Done()` is called.

The example below shows how to use Loop with a function which returns `error`.
This code reads CSV until Loop ends to EOF or an entry with Age < 0 is found in the CSV.

```golang
r.Loop(func(entry *struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}) error {
	fmt.Println(entry)
	if Age < 0 {
		return errors.New("Age mustn't be negative")
	}
})
if err := r.Done(); err != nil {
	log.Fatalf("Failed to read a CSV file: %v", err)
}
```

## ReadAll
```golang
func (r *Reader) ReadAll(s interface{})
```

ReadAll reads CSV to the end and convert all rows into the slice passed as an argument.
The argument `s` should be a pointer of a slice of a struct (`*[]myStruct`) or a pointer of a slice of a slice (`*[][]int`).
Aside from that, the same rules of Read are applied to ReadAll. You need to specify how to map columns to struct fields using struct field's tag.

```golang
var entry []struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}
r.ReadAll(&entry)
```

# Option
To control the behavior of Reader, you can pass Option to NewReader methods.

NewReader methods receive Option as a variadic parameter `opts`. `opts` is a variadic parameter so that we can omit `opts` from parameters when we call NewReader methods without changing Option.
Thus, you don't need to pass multiple Option to NewReader methods although you can pass as many Option as you want.

## Comma
Like [csv.Reader](https://golang.org/pkg/encoding/csv/#Reader) in the standard library, you can change the deliminator of CSV by specifying `Comma` option. For example, if you set `'\t'` to Comma, Reader reads a file as a TSV file.

## Comment
Comment, if not 0, is the comment character. Lines beginning with the character without preceding whitespace are ignored.

# Customizing decoders
By default, easycsv converts strings in CSV to integers, floats and bool automatically based on the types of struct fields and slices.

- Integers are parsed with `strconv.ParseInt` and unsigned integers are parsed with `strconv.ParseUint`.
  easycsv parses inputs as decimals by default. But it parses inputs as hex if inputs have `"0x"` prefix and
  as octal if inputs have `"0"` prefix (`"0xff"` → 255, `"077"` → 63).
- Floats are parsed with `strconv.ParseFloat`.
- bool is parsed with `strconv.ParseBool`.

You can customize how to decode strings in CSV to values by specifying `enc` attribute to struct fields.

## Predefined encoding
TBD

## Custom encoding
TBD

# godoc
[godoc](https://godoc.org/github.com/yunabe/easycsv)
