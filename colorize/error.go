package colorize

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func Red(v string) Part {
	return &partImpl{redColorFun, v, false}
}

func Cyan(v string) Part {
	return &partImpl{cyanColorFun, v, true}
}

func Green(v string) Part {
	return &partImpl{greenColorFun, v, false}
}

func None(v string) Part {
	return &partImpl{noColorFun, v, false}
}

var (
	redColorFun   = color.HiRedString
	cyanColorFun  = color.HiCyanString
	greenColorFun = color.HiGreenString
	noColorFun    = fmt.Sprintf
)

type Error struct {
	parts    []Part
	subError error
	postfix  []Part
}

func (e *Error) AddParts(values ...Part) *Error {
	if e.subError != nil {
		panic("can't add parts after set suberror")
	}
	e.parts = append(e.parts, values...)
	return e
}

func (e *Error) AddPostfix(values ...Part) *Error {
	e.postfix = append(e.postfix, values...)
	return e
}

func (e *Error) SetSubError(err error) *Error {
	e.subError = err
	return e
}

func (e *Error) Error() string {
	buf := strings.Builder{}
	for _, p := range e.parts {
		_, _ = buf.WriteString(p.Text())
	}
	if e.subError != nil {
		_, _ = buf.WriteString(": ")
		_, _ = buf.WriteString(e.subError.Error())
	}
	for _, p := range e.postfix {
		_, _ = buf.WriteString(p.Text())
	}
	return buf.String()
}

func (e *Error) ColorError() string {
	buf := strings.Builder{}
	for _, p := range e.parts {
		_, _ = buf.WriteString(p.ColorText())
	}
	if e.subError != nil {
		_, _ = buf.WriteString(": ")
		_, _ = buf.WriteString(GetColoredValue(e.subError))
	}
	for _, p := range e.postfix {
		_, _ = buf.WriteString(p.ColorText())
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
	if pErr, ok := err.(*Error); ok {
		return pErr.ColorError()
	}
	return err.Error()
}

func NewError(format string, values ...Part) *Error {
	plain := []Part{}
	for _, s := range strings.Split(format, "%s") {
		plain = append(plain, None(s))
	}
	return &Error{
		parts:   alternateJoin(plain, values),
		postfix: []Part{},
	}
}

func NewEntityError(pattern, entity string) *Error {
	return NewError(pattern, Cyan(entity))
}

func NewPathError(path string, subErr error) *Error {
	err := &Error{
		parts: []Part{None("path "), Cyan(path)},
	}
	return err.SetSubError(subErr)
}

func NewEntityNotEqualError(pattern, entity string, expected, actual interface{}) *Error {
	return NewError(
		pattern+"\n     expected: %s\n       actual: %s",
		Cyan(entity),
		Green(fmt.Sprintf("%v", expected)),
		Red(fmt.Sprintf("%v", actual)),
	)
}

func NewNotEqualError(pattern string, expected, actual interface{}) *Error {
	return NewError(
		pattern+"\n     expected: %s\n       actual: %s",
		Green(fmt.Sprintf("%v", expected)),
		Red(fmt.Sprintf("%v", actual)),
	)
}

func HasPathComponent(err error) bool {
	pErr, ok := err.(*Error)
	if !ok {
		return false
	}
	return len(pErr.parts) == 2 && pErr.parts[0].Text() == "path "
}

func RemovePathComponent(err error) error {
	if pErr, ok := err.(*Error); ok && HasPathComponent(err) {
		return pErr.subError
	}
	return err
}
