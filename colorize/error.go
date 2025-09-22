package colorize

import (
	"strings"
)

type Error struct {
	parts    []*Part
	subError error
	postfix  []*Part
}

func NewError(format string, values ...*Part) *Error {
	messageParts := strings.Split(format, "%s")

	lenMessage := len(messageParts)
	lenValues := len(values)

	lenMax := lenMessage
	if lenMax < lenValues {
		lenMax = lenValues
	}

	parts := make([]*Part, 0, lenMessage+lenValues)
	for i := 0; i < lenMax; i++ {
		if i < lenMessage {
			parts = append(parts, None(messageParts[i]))
		}
		if i < lenValues {
			parts = append(parts, values[i])
		}
	}

	return &Error{
		parts: parts,
	}
}

func (e *Error) Error() string {
	return ProcessWithTemplate(e, NoColorMap)
}

func (e *Error) Unwrap() error {
	return e.subError
}

func (e *Error) WithPostfix(values []*Part) *Error {
	e.postfix = values
	return e
}

func (e *Error) WithSubError(err error) *Error {
	e.subError = err
	return e
}

func GetColoredValue(err error) string {
	return ProcessWithTemplate(err, ColorizeMap)
}

func ProcessWithTemplate(err error, tpl map[Color]func(string) string) string {
	pErr, ok := err.(*Error)
	if !ok {
		return err.Error()
	}

	buf := &strings.Builder{}

	appendParts(buf, pErr.parts, tpl)

	if pErr.subError != nil {
		_, _ = buf.WriteString(": ")
		_, _ = buf.WriteString(ProcessWithTemplate(pErr.subError, tpl))
	}

	appendParts(buf, pErr.postfix, tpl)

	return buf.String()
}

func appendParts(buf *strings.Builder, parts []*Part, tpl map[Color]func(string) string) {
	for _, p := range parts {
		value := p.Value
		if f, ok := tpl[p.Color]; ok {
			value = f(value)
		}
		_, _ = buf.WriteString(value)
	}
}
