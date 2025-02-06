package terminal

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"text/template"

	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/models"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

type ColorPolicy int

const (
	PolicyAuto ColorPolicy = iota
	PolicyForceColor
	PolicyForceNoColor
)

//go:embed template.txt
var resultTmpl string

type OutputOpts struct {
	Policy       ColorPolicy
	ShowSuccess  bool
	CustomWriter io.Writer
}

type Output struct {
	opts    OutputOpts
	fprintf func(w io.Writer, format string, a ...interface{}) (n int, err error)
}

func NewOutput(opts *OutputOpts) *Output {
	o := &Output{
		fprintf: fmt.Fprintf,
	}
	if opts != nil {
		o.opts = *opts
	}
	if o.opts.Policy == PolicyAuto {
		if color.NoColor {
			o.opts.Policy = PolicyForceNoColor
		} else {
			o.opts.Policy = PolicyForceColor
		}
	}

	if o.opts.Policy == PolicyForceColor {
		o.fprintf = color.New().Fprintf
	}

	if o.opts.CustomWriter == nil {
		o.opts.CustomWriter = colorable.NewColorableStdout()
	}

	return o
}

func (o *Output) Process(_ models.TestInterface, result *models.Result) error {
	if !result.Passed() || o.opts.ShowSuccess {
		text, err := renderResult(result, o.opts.Policy)
		if err != nil {
			return err
		}
		_, _ = o.fprintf(o.opts.CustomWriter, "%s", text)
	}

	return nil
}

func renderResult(result *models.Result, policy ColorPolicy) (string, error) {
	_, hasHeaders := result.Test.GetResponseHeaders(result.ResponseStatusCode)

	var buffer bytes.Buffer
	t := template.Must(template.New("report").Funcs(getTemplateFuncMap(policy, hasHeaders)).Parse(resultTmpl))
	if err := t.Execute(&buffer, result); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func getTemplateFuncMap(policy ColorPolicy, showHeaders bool) template.FuncMap {
	var funcMap template.FuncMap
	if policy == PolicyForceColor {
		funcMap = template.FuncMap{
			"green":      color.GreenString,
			"cyan":       color.CyanString,
			"yellow":     color.YellowString,
			"danger":     color.New(color.FgHiWhite, color.BgRed).Sprint,
			"success":    color.New(color.FgHiWhite, color.BgGreen).Sprint,
			"printError": colorize.GetColoredValue,
		}
	} else {
		funcMap = template.FuncMap{
			"green":      fmt.Sprintf,
			"cyan":       fmt.Sprintf,
			"yellow":     fmt.Sprintf,
			"danger":     fmt.Sprintf,
			"success":    fmt.Sprintf,
			"printError": suppressColor,
		}
	}
	funcMap["inc"] = func(i int) int { return i + 1 }
	funcMap["show_headers"] = func() bool { return showHeaders }
	return funcMap
}

func suppressColor(err error) string {
	return err.Error()
}
