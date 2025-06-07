package beans

import (
	"encoding/xml"
	"time"
)

const NsModel = `urn:model.allure.qatools.yandex.ru`

type Suite struct {
	XMLName   xml.Name `xml:"ns2:test-suite"`
	NsAttr    string   `xml:"xmlns:ns2,attr"`
	Start     int64    `xml:"start,attr"`
	End       int64    `xml:"stop,attr"`
	Name      string   `xml:"name"`
	Title     string   `xml:"title"`
	TestCases struct {
		Cases []*TestCase `xml:"test-case"`
	} `xml:"test-cases"`
}

func NewSuite(name string, start time.Time) *Suite {
	return &Suite{
		NsAttr: NsModel,
		Name:   name,
		Title:  name,
		Start:  microSeconds(start),
	}
}

// SetEnd set end time for suite
func (s *Suite) SetEnd(endTime time.Time) {
	s.End = microSeconds(endTime)
}

// suite has test-cases?
func (s *Suite) HasTests() bool {
	return len(s.TestCases.Cases) > 0
}

// add test in suite
func (s *Suite) AddTest(test *TestCase) {
	s.TestCases.Cases = append(s.TestCases.Cases, test)
}
