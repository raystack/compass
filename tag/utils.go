package tag

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/odpf/columbus/tag/validator"
)

func buildFieldError(key string, message string) error {
	return ValidationError{
		validator.FieldError{
			key: message,
		},
	}
}

func ParseTagValue(templateURN string, fieldID uint,
	dataType string, tagValue string, options []string) (interface{}, error,
) {
	if tagValue == "" {
		return nil, nil
	}
	var output interface{}
	var err error
	switch dataType {
	case "double":
		output, err = strconv.ParseFloat(tagValue, 64)
		if err != nil {
			err = fmt.Errorf("template [%s] on field [%d] should be double", templateURN, fieldID)
		}
	case "boolean":
		output, err = strconv.ParseBool(tagValue)
		if err != nil {
			err = fmt.Errorf("template [%s] on field [%d] should be boolean", templateURN, fieldID)
		}
	case "enumerated":
		isValueValid := false
		for _, opt := range options {
			if tagValue == opt {
				isValueValid = true
				output = opt
				break
			}
		}
		if !isValueValid {
			err = fmt.Errorf("template [%s] on field [%d] should be one of (%s)",
				templateURN, fieldID, strings.Join(options, ", "),
			)
		}
	case "datetime":
		output, err = time.Parse(time.RFC3339, tagValue)
		if err != nil {
			err = fmt.Errorf("template [%s] on field [%d] should follow RFC3339",
				templateURN, fieldID,
			)
		}
	case "string":
		output = tagValue
	}
	return output, err
}
