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
