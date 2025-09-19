package colorize

type Part interface {
	Text() string
	ColorText() string
}

type colorerFun func(format string, a ...interface{}) string

type partImpl struct {
	colorer colorerFun
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
	return p.colorer("%s", p.value)
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
