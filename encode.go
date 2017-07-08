package easycsv

import (
	"fmt"
	"reflect"
	"strconv"
)

var predefinedDecoders = map[string]func(t reflect.Type) interface{}{
	"hex": func(t reflect.Type) interface{} {
		return createIntConverter(t, 16)
	},
	"oct": func(t reflect.Type) interface{} {
		return createIntConverter(t, 8)
	},
	"deci": func(t reflect.Type) interface{} {
		return createIntConverter(t, 10)
	},
}

func createIntConverter(t reflect.Type, base int) interface{} {
	switch t.Kind() {
	case reflect.Int:
		return func(s string) (int, error) {
			i, err := strconv.ParseInt(s, base, 0)
			return int(i), err
		}
	case reflect.Int8:
		return func(s string) (int8, error) {
			i, err := strconv.ParseInt(s, base, 8)
			return int8(i), err
		}
	case reflect.Int16:
		return func(s string) (int16, error) {
			i, err := strconv.ParseInt(s, base, 16)
			return int16(i), err
		}
	case reflect.Int32:
		return func(s string) (int32, error) {
			i, err := strconv.ParseInt(s, base, 32)
			return int32(i), err
		}
	case reflect.Int64:
		return func(s string) (int64, error) {
			i, err := strconv.ParseInt(s, base, 64)
			return int64(i), err
		}
	case reflect.Uint:
		return func(s string) (uint, error) {
			i, err := strconv.ParseUint(s, base, 0)
			return uint(i), err
		}
	case reflect.Uint8:
		return func(s string) (uint8, error) {
			i, err := strconv.ParseUint(s, base, 8)
			return uint8(i), err
		}
	case reflect.Uint16:
		return func(s string) (uint16, error) {
			i, err := strconv.ParseUint(s, base, 16)
			return uint16(i), err
		}
	case reflect.Uint32:
		return func(s string) (uint32, error) {
			i, err := strconv.ParseUint(s, base, 32)
			return uint32(i), err
		}
	case reflect.Uint64:
		return func(s string) (uint64, error) {
			i, err := strconv.ParseUint(s, base, 32)
			return uint64(i), err
		}
	default:
		return nil
	}
}

func validateTypeDecoder(t reflect.Type, conv interface{}) error {
	convT := reflect.TypeOf(conv)
	if convT.Kind() != reflect.Func {
		return fmt.Errorf("The decoder for %v must be a function but %v", t, convT)
	}
	if convT.NumIn() != 1 || convT.NumOut() != 2 {
		return fmt.Errorf("The decoder for %v must receive one arguments and returns two values", t)
	}
	if convT.In(0).Kind() != reflect.String {
		return fmt.Errorf("The decoder for %v must receive a string as the first arg, but receives %v", t, convT.In(0))
	}
	if convT.Out(0) != t || convT.Out(1) != errorType {
		return fmt.Errorf("The decoder for %v must return (%v, error), but returned (%v, %v)",
			t, t, convT.Out(0), convT.Out(1))
	}
	return nil
}

func createConverterFromType(opt Option, t reflect.Type) (interface{}, error) {
	if opt.TypeDecoders != nil {
		if conv, ok := opt.TypeDecoders[t]; ok {
			if err := validateTypeDecoder(t, conv); err != nil {
				return nil, err
			}
			return conv, nil
		}
	}
	return createDefaultConverter(t), nil
}

func createDefaultConverter(t reflect.Type) interface{} {
	c := createIntConverter(t, 0)
	if c != nil {
		return c
	}
	switch t.Kind() {
	case reflect.Float32:
		return func(s string) (float32, error) {
			f, err := strconv.ParseFloat(s, 32)
			return float32(f), err
		}
	case reflect.Float64:
		return func(s string) (float64, error) {
			f, err := strconv.ParseFloat(s, 64)
			return float64(f), err
		}
	case reflect.Bool:
		return strconv.ParseBool
	case reflect.String:
		return func(s string) (string, error) {
			return s, nil
		}
	default:
		return nil
	}
}
