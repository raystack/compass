package cli

import (
	"errors"

	"github.com/MakeNowJust/heredoc"
)

var (
	ErrConfigNotFound = errors.New(heredoc.Doc(`
	Config file not found. Loading from defaults...

	Run "compass config init" to initialize a new configuartion file 
	Run "compass help environment" for more information.

	Alternatively, make a "compass.yaml" file in the current directory from the example given
`))
)
