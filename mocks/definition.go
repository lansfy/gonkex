package mocks

import (
	"errors"
	"net/http"
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
	order              *orderChecker
	orderValue         int
}

func NewDefinition(path string, constraints []verifier, strategy ReplyStrategy, callsConstraint int, orderValue int) *Definition {
	return &Definition{
		path:               path,
		requestConstraints: constraints,
		replyStrategy:      strategy,
		callsConstraint:    callsConstraint,
		orderValue:         orderValue,
	}
}

func (d *Definition) Execute(w http.ResponseWriter, r *http.Request) []error {
	var err error

	d.mutex.Lock()
	d.calls++
	d.mutex.Unlock()

	if d.order != nil {
		err = d.order.Update(d.orderValue)
	}

	errs := verifyRequestConstraints(d.requestConstraints, r)
	if d.replyStrategy != nil {
		errs = append(errs, d.replyStrategy.HandleRequest(w, r)...)
	}
	if err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (d *Definition) ResetRunningContext() {
	if s, ok := d.replyStrategy.(contextAwareStrategy); ok {
		s.ResetRunningContext()
	}

	d.mutex.Lock()
	d.calls = 0
	if d.order != nil {
		d.order.Reset()
	}
	d.mutex.Unlock()
}

func (d *Definition) EndRunningContext(intermediate bool) []error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var errs []error
	if s, ok := d.replyStrategy.(contextAwareStrategy); ok {
		errs = s.EndRunningContext(intermediate)
	}

	if intermediate {
		return errs
	}

	if d.order != nil {
		d.order.Reset()
	}

	if d.callsConstraint != CallsNoConstraint && d.calls != d.callsConstraint {
		errs = append(errs, colorize.NewEntityError("path %s", d.path).SetSubError(
			colorize.NewNotEqualError("number of %s does not match:", "calls", d.callsConstraint, d.calls),
		))
	}
	return errs
}

func verifyRequestConstraints(requestConstraints []verifier, r *http.Request) []error {
	if len(requestConstraints) == 0 {
		return []error{}
	}

	var dump colorize.Part
	var errs []error
	for _, c := range requestConstraints {
		for _, e := range c.Verify(r) {
			if dump == nil {
				dump = colorize.None(dumpRequest(r))
			}
			errs = append(errs, colorize.NewEntityError("request constraint %s", c.GetName()).SetSubError(e).AddParts(
				colorize.None(", request was:\n\n"), dump,
			))
		}
	}

	return errs
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
