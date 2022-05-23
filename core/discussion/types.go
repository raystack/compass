package discussion

type Type string

const (
	TypeOpenEnded Type = "openended"
	TypeIssues    Type = "issues"
	TypeQAndA     Type = "qanda"
)

var SupportedTypes = []string{TypeOpenEnded.String(), TypeIssues.String(), TypeQAndA.String()}

// String returns type enum as string
func (dt Type) String() string {
	return string(dt)
}

// GetTypeEnum converts string to type enum
func GetTypeEnum(dt string) Type {
	switch {
	case dt == TypeOpenEnded.String():
		return TypeOpenEnded
	case dt == TypeIssues.String():
		return TypeIssues
	case dt == TypeQAndA.String():
		return TypeQAndA
	}
	// fallback
	return TypeOpenEnded
}

// IsTypeStringValid returns true if type string is valid/supported
func IsTypeStringValid(ts string) bool {
	for _, supported := range SupportedTypes {
		if supported == ts {
			return true
		}
	}
	return false
}
