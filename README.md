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

[`Read`](https://godoc.org/github.com/yunabe/easycsv#Reader.Read) receives a pointer to a struct (e.g. `*mystruct`) or a pointer to a slice of a primitive type (e.g. `*[]int`).
If it reads a new row from CSV successufly, it stores the row into `e` and returns `true`.
If `Reader` reached to `EOF` or it fails to read a new row for some reasons, it returns `false`.

`Read` can return `false` for a lot of reasons. To know the reason, you have to call `Done()` subsequently.
`Done` returns an error if `Read` encountered an error.
`Done` returns `nil` if `Read` returned `false` because it reached to `EOF`.

## Loop
TBD

## ReadAll
TBD

# godoc
[godoc](https://godoc.org/github.com/yunabe/easycsv)