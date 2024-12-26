package terminal

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/fatih/color"
	"github.com/lansfy/gonkex/models"
)

type ColorPolicy int

const (
	PolicyAuto ColorPolicy = iota
	PolicyForceColor
	PolicyForceNoColor
)

const dotsPerLine = 80

//go:embed template.txt
var resultTmpl string

type TerminalOutputOpts struct {
	Policy       ColorPolicy
	ShowSuccess  bool
	ShowProgress bool
}

type TerminalOutput struct {
	opts   TerminalOutputOpts
	printf func(format string, a ...interface{})
	dots   int
}

func NewOutput(opts *TerminalOutputOpts) *TerminalOutput {
	o := &TerminalOutput{
		printf: printf,
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
		o.printf = color.New().PrintfFunc()
	}

	return o
}

func (o *TerminalOutput) Process(_ models.TestInterface, result *models.Result) error {
	if !result.Passed() || o.opts.ShowSuccess {
		text, err := renderResult(result, o.opts.Policy)
		if err != nil {
			return err
		}
		o.printf("%s", text)
	} else if o.opts.ShowProgress {
		o.printf(".")
		o.dots++
		if o.dots%dotsPerLine == 0 {
			o.printf("\n")
		}
	}

	return nil
}

func renderResult(result *models.Result, policy ColorPolicy) (string, error) {
	var buffer bytes.Buffer
	t := template.Must(template.New("report").Funcs(getTemplateFuncMap(policy)).Parse(resultTmpl))
	if err := t.Execute(&buffer, result); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (o *TerminalOutput) ShowSummary(summary *models.Summary) {
	o.printf(
		"\nsuccess %d, failed %d, skipped %d, broken %d, total %d\n",
		summary.Total-summary.Broken-summary.Failed-summary.Skipped,
		summary.Failed,
		summary.Skipped,
		summary.Broken,
		summary.Total,
	)
}

func printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func getTemplateFuncMap(policy ColorPolicy) template.FuncMap {
	if policy == PolicyForceColor {
		return template.FuncMap{
			"green":   color.GreenString,
			"cyan":    color.CyanString,
			"yellow":  color.YellowString,
			"danger":  color.New(color.FgHiWhite, color.BgRed).Sprint,
			"success": color.New(color.FgHiWhite, color.BgGreen).Sprint,
			"inc":     func(i int) int { return i + 1 },
		}
	}
	return template.FuncMap{
		"green":   fmt.Sprintf,
		"cyan":    fmt.Sprintf,
		"yellow":  fmt.Sprintf,
		"danger":  fmt.Sprintf,
		"success": fmt.Sprintf,
		"inc":     func(i int) int { return i + 1 },
	}
}
