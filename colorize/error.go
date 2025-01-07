package colorize

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func Red(v interface{}) Part {
	return &partImpl{color.HiRedString, fmt.Sprintf("%v", v), false}
}

func Cyan(v interface{}) Part {
	return &partImpl{color.HiCyanString, fmt.Sprintf("%v", v), true}
}

func Green(v interface{}) Part {
	return &partImpl{color.HiGreenString, fmt.Sprintf("%v", v), false}
}

func None(v interface{}) Part {
	return &partImpl{asIsString, fmt.Sprintf("%v", v), false}
}

func SubError(err error) Part {
	return &subErrorImpl{err}
}

type Error struct {
	parts []Part
}

func (e *Error) AddParts(values ...Part) *Error {
	e.parts = append(e.parts, values...)
	return e
}

func (e *Error) SetSubError(err error) *Error {
	return e.AddParts(SubError(err))
}

func (e *Error) Error() string {
	buf := strings.Builder{}
	for _, p := range e.parts {
		buf.WriteString(p.Text())
	}
	return buf.String()
}

func asIsString(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func (e *Error) ColorError() string {
	buf := strings.Builder{}
	for _, p := range e.parts {
		buf.WriteString(p.ColorText())
	}
	return buf.String()
}

func alternateJoin(list1, list2 []Part) []Part {
	result := []Part{}
	i, j := 0, 0
	for len(result) != len(list1)+len(list2) {
		if i < len(list1) {
			result = append(result, list1[i])
			i++
		}
		if j < len(list2) {
			result = append(result, list2[j])
			j++
		}
	}

	return result
}

func GetColoredValue(err error) string {
	var pErr *Error
	if errors.As(err, &pErr) {
		return pErr.ColorError()
	}
	return err.Error()
}

func NewError(format string, values ...Part) *Error {
	plain := []Part{}
	for _, s := range strings.Split(format, "%s") {
		plain = append(plain, None(s))
	}
	return &Error{alternateJoin(plain, values)}
}

func NewEntityError(pattern, entity string) *Error {
	return NewError(pattern, Cyan(entity))
}

func NewNotEqualError(pattern, entity string, expected, actual interface{}) *Error {
	pattern += "\n     expected: %s\n       actual: %s"
	return NewError(pattern, Cyan(entity), Green(expected), Red(actual))
}

func NewNotEqualError2(before, entity, after string, expected, actual interface{}) *Error {
	parts := []Part{
		None(before), Cyan(entity), None(after + "\n     expected: "), Green(expected), None("\n       actual: "), Red(actual),
	}
	return NewError("", parts...)
}
