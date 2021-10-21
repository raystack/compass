package store

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/columbus/models"
	"github.com/pkg/errors"
)

func mapToV2(src map[string]interface{}) (record models.RecordV2, err error) {
	err = decode(src, &record)
	if err != nil {
		err = errors.Wrap(err, "error decoding data")
		return
	}

	isV2Record := record.Data != nil && record.Name != "" && record.Urn != ""
	if !isV2Record {
		err = errors.New("record is not in a v2 format")
		return
	}

	return
}

// decode is being used to map v2 record to v1 record
func decode(input map[string]interface{}, result interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: result,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(input); err != nil {
		return err
	}
	return err
}

func toTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}
