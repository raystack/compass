package v1beta1

import (
	"errors"
	"fmt"
)

var (
	errMissingUserInfo = errors.New("missing user information")
)

func bodyParserErrorMsg(err error) string {
	return fmt.Sprintf("error parsing request body: %v", err)
}
