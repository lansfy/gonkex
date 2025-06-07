package beans

import (
	"time"
)

type Step struct {
	Parent *Step `xml:"-"`

	Status      string        `xml:"status,attr"`
	Start       int64         `xml:"start,attr"`
	Stop        int64         `xml:"stop,attr"`
	Name        string        `xml:"name"`
	Steps       []*Step       `xml:"steps"`
	Attachments []*Attachment `xml:"attachments"`
}

func NewStep(name string, start time.Time) *Step {
	return &Step{
		Name:  name,
		Start: microSeconds(start),
	}
}

func (s *Step) End(status string, end time.Time) {
	s.Status = status
	s.Stop = microSeconds(end)
}

func (s *Step) AddStep(step *Step) {
	if step != nil {
		s.Steps = append(s.Steps, step)
	}
}

func microSeconds(t time.Time) int64 {
	return t.UnixNano() / 1000
}
