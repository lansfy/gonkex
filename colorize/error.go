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

func SubError(err error) Part {
	return &subErrorImpl{err}
}

var (
	redColorFun   = color.HiRedString
	cyanColorFun  = color.HiCyanString
	greenColorFun = color.HiGreenString
	noColorFun    = fmt.Sprintf
)

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
		_, _ = buf.WriteString(p.Text())
	}
	return buf.String()
}

func (e *Error) ColorError() string {
	buf := strings.Builder{}
	for _, p := range e.parts {
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
	return &Error{alternateJoin(plain, values)}
}

func NewEntityError(pattern, entity string) *Error {
	return NewError(pattern, Cyan(entity))
}

func NewNotEqualError(pattern, entity string, expected, actual interface{}) *Error {
	pattern += "\n     expected: %s\n       actual: %s"
	return NewError(pattern, Cyan(entity), Green(fmt.Sprintf("%v", expected)), Red(fmt.Sprintf("%v", actual)))
}

// TODO: remove this hack
func HasPathComponent(err error) bool {
	pErr, ok := err.(*Error)
	if !ok {
		return false
	}
	return pErr.parts[0].Text() == "path " && len(pErr.parts) >= 3
}

func RemovePathComponent(err error) error {
	pErr, ok := err.(*Error)
	if !ok {
		return err
	}
	if pErr.parts[0].Text() == "path " && len(pErr.parts) >= 3 {
		parts := pErr.parts
		for len(parts) != 0 {
			if strings.HasPrefix(parts[0].Text(), ": ") {
				parts[0] = None(parts[0].Text()[2:])
				break
			}
			parts = parts[1:]
		}
		pErr.parts = parts
	}
	return pErr
}
