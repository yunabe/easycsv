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

The conversion from CSV row (string) to the given field type (int, float32, bool, etc...) is handled in Reader automatically.

When you read CSV with `Read` methods, you have to always call `Done()` subsequently to (1) check the error and (2) close the file behind the Reader when it is instantiated with `NewReadCloser` or `NewReaderFile`. If you forget to call `Done()`, the error will be completely gone.

## Loop
```golang
func (r *Reader) Loop(body interface{})
```

## ReadAll
```golang
func (r *Reader) ReadAll(s interface{})
```

ReadAll reads CSV to the end and convert all rows into the slice passed as an argument.
The argument `s` should be a pointer of a slice of a struct (`*[]myStruct`) or a pointer of a slice of a slice (`*[][]int`).
Aside from that, the same rule of ReadAll is applied to ReadAll. You need to specify how to map columns to struct fields using struct field's tag.

```golang
var entry []struct {
	Name string `index:"0"`
	Age  int    `index:"1"`
}
r.ReadAll(&entry)
```

# godoc
[godoc](https://godoc.org/github.com/yunabe/easycsv)
