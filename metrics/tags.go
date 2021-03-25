package metrics

import "fmt"

type Tags struct {
	RequestMethod string
	RequestUrl    string
}

func (m Tags) String() string {
	return fmt.Sprintf("method=%s,url=%s", m.RequestMethod, m.RequestUrl)
}
