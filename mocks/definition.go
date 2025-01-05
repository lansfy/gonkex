package mocks

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/lansfy/gonkex/colorize"
)

var _ contextAwareStrategy = (*Definition)(nil)

const CallsNoConstraint = -1

type Definition struct {
	path               string
	requestConstraints []verifier
	replyStrategy      ReplyStrategy
	mutex              sync.Mutex
	calls              int
	callsConstraint    int
}

func NewDefinition(path string, constraints []verifier, strategy ReplyStrategy, callsConstraint int) *Definition {
	return &Definition{
		path:               path,
		requestConstraints: constraints,
		replyStrategy:      strategy,
		callsConstraint:    callsConstraint,
	}
}

func (d *Definition) Execute(w http.ResponseWriter, r *http.Request) []error {
	d.mutex.Lock()
	d.calls++
	d.mutex.Unlock()

	errors := verifyRequestConstraints(d.requestConstraints, r)
	if d.replyStrategy != nil {
		errors = append(errors, d.replyStrategy.HandleRequest(w, r)...)
	}
	return errors
}

func (d *Definition) ResetRunningContext() {
	if s, ok := d.replyStrategy.(contextAwareStrategy); ok {
		s.ResetRunningContext()
	}
	d.mutex.Lock()
	d.calls = 0
	d.mutex.Unlock()
}

func (d *Definition) EndRunningContext() []error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	var errs []error
	if s, ok := d.replyStrategy.(contextAwareStrategy); ok {
		errs = s.EndRunningContext()
	}
	if d.callsConstraint != CallsNoConstraint && d.calls != d.callsConstraint {
		err := colorize.NewEntityError("at path %s", d.path)
		err.SetSubError(
			colorize.NewNotEqualError("number of %s does not match:", "calls", d.callsConstraint, d.calls),
		)
		errs = append(errs, err)
	}
	return errs
}

func verifyRequestConstraints(requestConstraints []verifier, r *http.Request) []error {
	if len(requestConstraints) == 0 {
		return []error{}
	}

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		requestDump = []byte(fmt.Sprintf("dump request: %v", err))
	}

	var errors []error
	for _, c := range requestConstraints {
		errs := c.Verify(r)
		for _, e := range errs {
			err := colorize.NewEntityError("request constraint %s", c.GetName()).SetSubError(e)
			err.AddParts(colorize.None(", request was:\n\n"), colorize.None(string(requestDump)))
			errors = append(errors, err)
		}
	}

	return errors
}
func (d *Definition) ExecuteWithoutVerifying(w http.ResponseWriter, r *http.Request) []error {
	d.mutex.Lock()
	d.calls++
	d.mutex.Unlock()
	if d.replyStrategy != nil {
		return d.replyStrategy.HandleRequest(w, r)
	}
	return []error{
		errors.New("reply strategy undefined"),
	}
}
