# easycsv
easycsv package provides API to read CSV file in Go (golang).

# Installation
```
go get -u github.com/yunabe/golang-codelab/easycsv
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
TBD

# gdoc
[godoc](https://godoc.org/github.com/yunabe/easycsv)