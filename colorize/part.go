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
	return p.colorer("%s", p.value)
}
