package colorize

type Color int

const (
	ColorNone Color = iota
	ColorRed
	ColorCyan
	ColorGreen
)

type Part struct {
	Color Color
	Value string
}

func Red(v string) *Part {
	return &Part{ColorRed, v}
}

func Cyan(v string) *Part {
	return &Part{ColorCyan, v}
}

func Green(v string) *Part {
	return &Part{ColorGreen, v}
}

func None(v string) *Part {
	return &Part{ColorNone, v}
}
