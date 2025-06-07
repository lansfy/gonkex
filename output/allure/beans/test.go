package beans

import (
	"strings"
	"time"
)

type TestCase struct {
	Status string `xml:"status,attr"`
	Start  int64  `xml:"start,attr"`
	Stop   int64  `xml:"stop,attr"`
	Name   string `xml:"name"`
	Steps  struct {
		Steps []*Step `xml:"step"`
	} `xml:"steps"`
	Labels struct {
		Label []*Label `xml:"label"`
	} `xml:"labels"`
	Attachments struct {
		Attachment []*Attachment `xml:"attachment"`
	} `xml:"attachments"`
	Desc    string `xml:"description"`
	Failure struct {
		Msg   string `xml:"message"`
		Trace string `xml:"stack-trace"`
	} `xml:"failure,omitempty"`
}

func NewTestCase(name string, start time.Time) *TestCase {
	return &TestCase{
		Name:  name,
		Start: microSeconds(start),
	}
}

func (t *TestCase) SetDescription(desc string) {
	t.Desc = desc
}

func (t *TestCase) AddLabel(name, value string) {
	t.Labels.Label = append(t.Labels.Label, NewLabel(name, value))
}

func (t *TestCase) AddStep(step *Step) {
	t.Steps.Steps = append(t.Steps.Steps, step)
}

func (t *TestCase) AddAttachment(attach *Attachment) {
	t.Attachments.Attachment = append(t.Attachments.Attachment, attach)
}

func (t *TestCase) End(status string, err error, end time.Time) {
	t.Status = status
	t.Stop = microSeconds(end)
	if err != nil {
		msg := strings.Split(err.Error(), "\trace")
		t.Failure.Msg = msg[0]
		t.Failure.Trace = strings.Join(msg[1:], "\n")
	}
}
