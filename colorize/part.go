package colorize

type Part interface {
	Text() string
	ColorText() string
}

type partImpl struct {
	colorer func(format string, a ...interface{}) string
	value   string
	entity  bool
}

func (p *partImpl) Text() string {
	if p.entity {
		return "'" + p.value + "'"
	}
	return p.value
}

func (p *partImpl) ColorText() string {
	return p.colorer(p.value)
}

type subErrorImpl struct {
	err error
}

func (p *subErrorImpl) Text() string {
	return ": " + p.err.Error()
}

func (p *subErrorImpl) ColorText() string {
	return ": " + GetColoredValue(p.err)
}
