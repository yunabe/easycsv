package easycsv

import (
	"strings"
	"testing"
	"time"
)

func TestIssue1(t *testing.T) {
	// https://github.com/yunabe/easycsv/issues/1
	// TODO(yunabe): Fix this bug.
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
	expectedErr := "\"date\" is not defined"
	err := r.Done()
	if err == nil {
		t.Errorf("Expected %v but got no error", expectedErr)
	}
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected an error with %q but got %q", expectedErr, err)
	}
}
