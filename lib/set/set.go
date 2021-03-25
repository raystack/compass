package set

import "encoding/json"

// StringSet is a lightweight set implementation for strings
// produces the correct representation for sets in JSON (string lists)
// while also allow'ing (set) map-like access in code.
// Same ordering guarantee's as Go's map type
type StringSet map[string]bool

func (ss StringSet) MarshalJSON() ([]byte, error) {
	var values []string
	for v := range ss {
		values = append(values, v)
	}
	return json.Marshal(values)
}

func (ss StringSet) Add(v string) StringSet {
	ss[v] = true
	return ss
}

func (ss *StringSet) UnmarshalJSON(data []byte) error {
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	ss.truncate()
	if *ss == nil {
		*ss = make(StringSet)
	}
	for _, value := range values {
		ss.Add(value)
	}
	return nil
}

func (ss StringSet) truncate() {
	for v := range ss {
		delete(ss, v)
	}
}

func NewStringSet(values ...string) StringSet {
	ss := make(StringSet)
	for _, value := range values {
		ss.Add(value)
	}
	return ss
}
