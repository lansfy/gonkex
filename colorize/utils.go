package colorize

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
)

func NewEntityError(pattern, entity string) *Error {
	return NewError(pattern, Cyan(entity))
}

func NewNotEqualError(pattern string, expected, actual interface{}) *Error {
	return NewError(
		pattern+"\n     expected: %s\n       actual: %s",
		Green(fmt.Sprintf("%v", expected)),
		Red(fmt.Sprintf("%v", actual)),
	)
}

func NewEntityNotEqualError(pattern, entity string, expected, actual interface{}) *Error {
	return NewError(
		pattern+"\n     expected: %s\n       actual: %s",
		Cyan(entity),
		Green(fmt.Sprintf("%v", expected)),
		Red(fmt.Sprintf("%v", actual)),
	)
}

func NewPathError(path string, subErr error) *Error {
	err := &Error{
		parts: []*Part{None("path "), Cyan(path)},
	}
	return err.WithSubError(subErr)
}

func HasPathComponent(err error) bool {
	pErr, ok := err.(*Error)
	return ok && len(pErr.parts) == 2 && pErr.parts[0].Value == "path "
}

func RemovePathComponent(err error) error {
	if HasPathComponent(err) {
		err = errors.Unwrap(err)
	}
	return err
}

var (
	redColorFun   = color.HiRedString
	cyanColorFun  = color.HiCyanString
	greenColorFun = color.HiGreenString

	ColorizeMap = map[Color]func(string) string{
		ColorRed: func(val string) string {
			return redColorFun("%s", val)
		},
		ColorCyan: func(val string) string {
			return cyanColorFun("%s", val)
		},
		ColorGreen: func(val string) string {
			return greenColorFun("%s", val)
		},
	}

	NoColorMap = map[Color]func(string) string{
		ColorCyan: func(val string) string {
			return "'" + val + "'"
		},
	}
)
