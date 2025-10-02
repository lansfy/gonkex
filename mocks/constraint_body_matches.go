package mocks

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/types"
)

func loadBodyMatchesConstraint(def map[string]interface{}, bodyType types.BodyType) (verifier, error) {
	body, err := getRequiredStringKey(def, "body", true)
	if err != nil {
		return nil, err
	}
	params, err := readCompareParams(def)
	if err != nil {
		return nil, err
	}

	return newBodyMatchesConstraint(body, params, bodyType)
}

func newBodyMatchesConstraint(expected string, params compare.Params, bodyType types.BodyType) (verifier, error) {
	expectedBody, err := bodyType.Decode(expected)
	if err != nil {
		return nil, fmt.Errorf("parse 'body': %w", err)
	}
	return &bodyMatchesConstraint{
		bodyType:      bodyType,
		expectedBody:  expectedBody,
		compareParams: params,
	}, nil
}

type bodyMatchesConstraint struct {
	bodyType      types.BodyType
	expectedBody  interface{}
	compareParams compare.Params
}

func (c *bodyMatchesConstraint) GetName() string {
	return "bodyMatches" + c.bodyType.GetName()
}

func (c *bodyMatchesConstraint) Verify(r *http.Request) []error {
	body, err := getRequestBodyCopy(r)
	if err != nil {
		return []error{err}
	}

	if len(body) == 0 {
		return []error{errors.New("request is empty")}
	}
	actual, err := c.bodyType.Decode(string(body))
	if err != nil {
		return []error{err}
	}
	return compare.Compare(c.expectedBody, actual, c.compareParams)
}
