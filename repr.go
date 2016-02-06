// Package repr attempts to represent Go values in a form that can be copy-and-pasted into source
// code directly.
//
// Unfortunately some values (such as pointers to basic types) can not be represented directly in
// Go. These values will be represented as `&<value>`. eg. `&23`
package repr

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
)

type reprOptions struct {
	indent string
}

func (r *reprOptions) nextIndent(indent string) string {
	if r.indent != "" {
		return indent + r.indent
	}
	return ""
}

func (r *reprOptions) thisIndent(indent string) string {
	if r.indent != "" {
		return indent
	}
	return ""
}

// Option modifies the default behaviour.
type Option func(o *reprOptions)

// Indent output by this much.
func Indent(indent string) Option { return func(o *reprOptions) { o.indent = indent } }

// Repr returns a string representing v.
func Repr(v interface{}, options ...Option) string {
	w := bytes.NewBuffer(nil)
	Write(w, v, options...)
	return w.String()
}

// Print v to os.Stdout on one line.
func Print(v interface{}, options ...Option) {
	Write(os.Stdout, v, options...)
	fmt.Fprintln(os.Stdout)
}

// Write writes a representation of v to w.
func Write(w io.Writer, v interface{}, options ...Option) {
	os := &reprOptions{}
	for _, o := range options {
		o(os)
	}
	reprValue(w, reflect.ValueOf(v), os, "")
}

func reprValue(w io.Writer, v reflect.Value, options *reprOptions, indent string) {
	in := options.thisIndent(indent)
	ni := options.nextIndent(indent)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		fmt.Fprintf(w, "%s{", v.Type())
		if options.indent != "" && v.Len() != 0 {
			fmt.Fprintf(w, "\n")
		}
		for i := 0; i < v.Len(); i++ {
			e := v.Index(i)
			fmt.Fprintf(w, "%s", ni)
			reprValue(w, e, options, ni)
			if options.indent != "" {
				fmt.Fprintf(w, ",\n")
			} else if i < v.Len()-1 {
				fmt.Fprintf(w, ", ")
			}
		}
		fmt.Fprintf(w, "%s}", in)
	case reflect.Chan:
		fmt.Fprintf(w, "make(")
		fmt.Fprintf(w, "%s", v.Type())
		fmt.Fprintf(w, ", %d)", v.Cap())
	case reflect.Map:
		fmt.Fprintf(w, "%s{", v.Type())
		if options.indent != "" && v.Len() != 0 {
			fmt.Fprintf(w, "\n")
		}
		for i, k := range v.MapKeys() {
			kv := v.MapIndex(k)
			fmt.Fprintf(w, "%s", ni)
			reprValue(w, k, options, ni)
			fmt.Fprintf(w, ": ")
			reprValue(w, kv, options, in)
			if options.indent != "" {
				fmt.Fprintf(w, ",\n")
			} else if i < v.Len()-1 {
				fmt.Fprintf(w, ", ")
			}
		}
		fmt.Fprintf(w, "%s}", in)
	case reflect.Struct:
		fmt.Fprintf(w, "%s{", v.Type())
		if options.indent != "" && v.NumField() != 0 {
			fmt.Fprintf(w, "\n")
		}
		for i := 0; i < v.NumField(); i++ {
			t := v.Type().Field(i)
			f := v.Field(i)
			fmt.Fprintf(w, "%s%s: ", ni, t.Name)
			reprValue(w, f, options, ni)
			if options.indent != "" {
				fmt.Fprintf(w, ",\n")
			} else if i < v.NumField()-1 {
				fmt.Fprintf(w, ", ")
			}
		}
		fmt.Fprintf(w, "%s}", indent)
	case reflect.Ptr:
		if v.IsNil() {
			fmt.Fprintf(w, "nil")
			return
		}
		fmt.Fprintf(w, "&")
		reprValue(w, v.Elem(), options, indent)
	case reflect.String:
		fmt.Fprintf(w, "%q", v.Interface())
	case reflect.Interface:
		reprValue(w, v.Elem(), options, indent)
	default:
		fmt.Fprintf(w, "%v", v)
	}
}
