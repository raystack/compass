package validator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FieldError is error with key is field name and value is all errors for that field
type FieldError map[string]string

// Error returns error that represent the field error
func (d FieldError) Error() string {
	var errFields []string
	for field, value := range d {
		errFields = append(errFields, fmt.Sprintf("%s : %s", field, value))
	}
	output := fmt.Sprintf("error with [%s]", strings.Join(errFields, ", "))
	return output
}

// JSON converts field error into its JSON representation
func (d FieldError) JSON() []byte {
	output, err := json.Marshal(d)
	if err != nil {
		return []byte(err.Error())
	}
	return output
}
