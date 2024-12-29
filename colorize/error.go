package colorize

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type Color int

const (
	ColorRed Color = iota
	ColorCyan
	ColorGreen
	ColorNone
)

type Part struct {
	attr  Color
	value string
}

func Red(v interface{}) *Part {
	return &Part{ColorRed, fmt.Sprintf("%v", v)}
}

func Cyan(v interface{}) *Part {
	return &Part{ColorCyan, fmt.Sprintf("%v", v)}
}

func Green(v interface{}) *Part {
	return &Part{ColorGreen, fmt.Sprintf("%v", v)}
}

func None(v interface{}) *Part {
	return &Part{ColorNone, fmt.Sprintf("%v", v)}
}

type Error struct {
	parts []*Part
}

func (e *Error) Error() string {
	buf := strings.Builder{}
	for _, p := range e.parts {
		buf.WriteString(p.value)
	}
	return buf.String()
}

func asIsString(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func (e *Error) ColorError() string {
	buf := strings.Builder{}
	for _, p := range e.parts {
		if p.value == "" {
			continue
		}
		f := asIsString
		switch p.attr {
		case ColorRed:
			f = color.RedString
		case ColorCyan:
			f = color.CyanString
		case ColorGreen:
			f = color.GreenString
		}

		buf.WriteString(f(p.value))
	}
	return buf.String()
}

func NewError(parts ...*Part) error {
	return &Error{parts}
}

func NewNotEqualError(before, entity, after string, expected, actual interface{}, tail []*Part) error {
	parts := []*Part{
		None(before), Cyan(entity), None(after + "\n     expected: "), Green(expected), None("\n       actual: "), Red(actual),
	}
	parts = append(parts, tail...)
	return NewError(parts...)
}
