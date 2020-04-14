package easycsv

import (
	"errors"
	"reflect"
)

// Option specifies the spec of Reader.
type Option struct {
	// Comma is the field delimiter.
	// For exampoe, if '\t' is set to Comma, Reader reads files as TSV files.
	Comma rune
	// Comment, if not 0, is the comment character. Lines beginning with the character without preceding whitespace are ignored.
	Comment rune
	// Allow lazy parsing of quotes, default to false
	LazyQuotes bool
	// If true, Reader does not check the number of fields per record
	AllowMissingFields bool
	// Decoders is the map to define custom encodings.
	Decoders map[string]interface{}
	// Custom decoders to parse specific types.
	TypeDecoders map[reflect.Type]interface{}

	// TODO: Support AutoIndex
	AutoIndex bool
	// TODO: Support AutoName
	AutoName bool
}

func (a *Option) mergeOption(b Option) {
	if b.Comma != 0 {
		a.Comma = b.Comma
	}
	if b.Comment != 0 {
		a.Comment = b.Comment
	}
	if b.AutoIndex {
		a.AutoIndex = true
	}
	if b.AutoName {
		a.AutoName = true
	}
	if b.LazyQuotes {
		a.LazyQuotes = b.LazyQuotes
	}
	if b.AllowMissingFields {
		a.AllowMissingFields = true
	}
	if b.Decoders != nil {
		if a.Decoders == nil {
			a.Decoders = make(map[string]interface{})
		}
		for name, dec := range b.Decoders {
			a.Decoders[name] = dec
		}
	}
	if b.TypeDecoders != nil {
		if a.TypeDecoders == nil {
			a.TypeDecoders = make(map[reflect.Type]interface{})
		}
		for t, dec := range b.TypeDecoders {
			a.TypeDecoders[t] = dec
		}
	}
}

func (a *Option) validate() error {
	if a.AutoIndex && a.AutoName {
		return errors.New("You can not set both AutoIndex and AutoName to easycsv.Reader.")
	}
	return nil
}

func mergeOptions(opts []Option) (Option, error) {
	var opt Option
	for _, o := range opts {
		opt.mergeOption(o)
	}
	return opt, opt.validate()
}
