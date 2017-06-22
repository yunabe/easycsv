package easycsv

import "errors"

// Option specifies the spec of Reader.
type Option struct {
	Comma rune
	// Comment, if not 0, is the comment character. Lines beginning with the character without preceding whitespace are ignored.
	Comment  rune
	Decoders map[string]interface{}
	// TODO: Use AutoIndex
	AutoIndex bool
	// TODO: Use AutoName
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
	if b.Decoders != nil {
		if a.Decoders == nil {
			a.Decoders = make(map[string]interface{})
		}
		for name, dec := range b.Decoders {
			a.Decoders[name] = dec
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
