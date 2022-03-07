package discussion

type State string

const (
	StateOpen   State = "open"
	StateClosed State = "closed"
)

var SupportedStates = []string{StateOpen.String(), StateClosed.String()}

// String returns state enum as string
func (ds State) String() string {
	return string(ds)
}

// GetStateEnum converts string to state enum
func GetStateEnum(ds string) State {
	switch {
	case ds == StateOpen.String():
		return StateOpen
	case ds == StateClosed.String():
		return StateClosed
	}
	// fallback
	return StateOpen
}

// IsStateStringValid returns true if state string is valid/supported
func IsStateStringValid(ss string) bool {
	for _, supported := range SupportedStates {
		if supported == ss {
			return true
		}
	}
	return false
}
