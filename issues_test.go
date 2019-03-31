package easycsv

import (
	"testing"
	"time"
)

func TestIssue1Fixed(t *testing.T) {
	// https://github.com/yunabe/easycsv/issues/1
	r := NewReaderFile("testdata/issue1.csv", Option{
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
		t.Fatalf("Failed to read: %v", err)
	}
	wantNames := []string{"Alice", "Bob"}
	wantBirths := []string{"1980/12/30", "1975/06/09"}
	noDiff(t, "names", names, wantNames)
	noDiff(t, "births", births, wantBirths)
}
