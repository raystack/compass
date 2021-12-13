package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldErrorAsString(t *testing.T) {
	key1 := "key1 error"
	message1 := "message1 error"

	key2 := "key2 error"
	message2 := "message2 error"

	err := FieldError{
		key1: message1,
		key2: message2,
	}

	expectedString := "error with [key1 error : message1 error, key2 error : message2 error]"

	assert.Equal(t, expectedString, err.Error())
}
