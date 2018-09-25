package easycsv

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestIssue1(t *testing.T) {
	// https://github.com/yunabe/easycsv/issues/1
	r := NewReader(bytes.NewBufferString("name,birthday\nAlice,1980-12-30\nBob,1975-06-09"), Option{
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
	var names []string
	var births []string
	for r.Read(&entry) {
		names = append(names, entry.Name)
		births = append(births, entry.Birth.Format("2006/01/02"))
	}
	if err := r.Done(); err != nil {
		t.Error(err)
		return
	}
	nameExpects := []string{"Alice", "Bob"}
	birthExpects := []string{"1980/12/30", "1975/06/09"}
	if !reflect.DeepEqual(names, nameExpects) {
		t.Errorf("Expected %v but got %v", nameExpects, names)
	}
	if !reflect.DeepEqual(births, birthExpects) {
		t.Errorf("Expected %v but got %v", birthExpects, births)
	}
}
