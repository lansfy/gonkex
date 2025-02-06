package yaml_file

import (
	"encoding/json"
	"fmt"

	"github.com/lansfy/gonkex/models"
)

type StatusEnum struct {
	value models.Status
}

var validStatuses = map[models.Status]struct{}{
	models.StatusNone:    {},
	models.StatusFocus:   {},
	models.StatusBroken:  {},
	models.StatusSkipped: {},
}

func (s *StatusEnum) UnmarshalJSON(data []byte) error {
	unmarshal := func(v interface{}) error {
		return json.Unmarshal(data, v)
	}
	return s.UnmarshalYAML(unmarshal)
}

func (s *StatusEnum) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return fmt.Errorf("wrong type for status value")
	}

	if _, exists := validStatuses[models.Status(str)]; !exists {
		return fmt.Errorf("unsupported value for status: %s", str)
	}

	s.value = models.Status(str)
	return nil
}
