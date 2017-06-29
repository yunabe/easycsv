package easycsv

import (
	"fmt"
	"log"
)

func ExampleReader_read() {
	r := NewReaderFile("testdata/sample.csv")
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
	// Output: {Alice 10}{Bob 20}
}

func ExampleReader_loop() {
	r := NewReaderFile("testdata/sample.csv")
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
	// Output: &{Alice 10}&{Bob 20}
}

func ExampleReader_readAll() {
	r := NewReaderFile("testdata/sample.csv")
	var entry []struct {
		Name string `index:"0"`
		Age  int    `index:"1"`
	}
	r.ReadAll(&entry)
	if err := r.Done(); err != nil {
		log.Fatalf("Failed to read a CSV file: %v", err)
	}
	fmt.Println(entry)
	// Output: [{Alice 10} {Bob 20}]
}

func ExampleReader_tSV() {
	r := NewReaderFile("testdata/sample.tsv", Option{
		Comma: '\t',
	})
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
	// Output: {Alice 10}{Bob 20}
}

func ExampleReader_LineNumber_reader() {
	r := NewReaderFile("testdata/sample.csv")
	var entry struct {
		Name string `index:"0"`
		Age  int    `index:"1"`
	}
	bob := "Bob"
	lino := 0
	for r.Read(&entry) {
		if entry.Name == bob {
			lino = r.LineNumber()
		}
	}
	if lino > 0 {
		fmt.Printf("Found %s at line %d", bob, lino)
	}
	// Output: Found Bob at line 2
}

func ExampleReader_DoneDefer() {
	f := func() (err error) {
		r := NewReaderFile("testdata/sample.csv")
		defer r.DoneDefer(&err)
		var entry struct {
			Name string `index:"3"`
		}
		// This fails with a index-out-of-range error.
		for r.Read(&entry) {
			err = fmt.Errorf("This block must not be executed")
		}
		return
	}
	err := f()
	if err != nil {
		fmt.Printf("Failed: %v", err)
	}
	// Output: Failed: Accessed index 3 though the size of the row is 2
}
