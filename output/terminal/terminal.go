package terminal

import (
	"bytes"
	_ "embed"
	"encoding/json"
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
	Template     string
	TemplateFunc template.FuncMap
	Writer       io.Writer
	PrettyBody   bool
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

	if o.opts.Writer == nil {
		o.opts.Writer = colorable.NewColorableStdout()
	}

	if o.opts.Template == "" {
		o.opts.Template = resultTmpl
	}

	return o
}

func (o *Output) Process(_ models.TestInterface, result *models.Result) error {
	if !result.Passed() || o.opts.ShowSuccess {
		text, err := o.renderResult(result)
		if err != nil {
			return err
		}
		_, _ = o.fprintf(o.opts.Writer, "%s", text)
	}

	return nil
}

func (o *Output) renderResult(result *models.Result) (string, error) {
	var buffer bytes.Buffer
	t := template.Must(template.New("report").Funcs(o.getTemplateFuncMap()).Parse(o.opts.Template))
	if err := t.Execute(&buffer, result); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (o *Output) getTemplateFuncMap() template.FuncMap {
	var funcMap template.FuncMap
	if o.opts.Policy == PolicyForceColor {
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
	funcMap["prettify"] = func(body string) string {
		if !o.opts.PrettyBody {
			return body
		}
		return makePretty(body)
	}

	for name, f := range o.opts.TemplateFunc {
		funcMap[name] = f
	}

	return funcMap
}

func makePretty(body string) string {
	out := &bytes.Buffer{}
	if json.Indent(out, []byte(body), "", "  ") != nil {
		return body
	}
	return out.String()
}

func suppressColor(err error) string {
	return err.Error()
}
