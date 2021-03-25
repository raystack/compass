package metrics

import (
	"fmt"
	"net"
	"strconv"

	statsd "github.com/etsy/statsd/examples/go"
)

type Client interface {
	Timing(string, int64)
	Increment(string)
}

type Monitor struct {
	client    Client
	prefix    string
	separator string
}

func (mm *Monitor) ResponseTime(requestMethod string, requestUrl string, responseTime int64) {
	statName := fmt.Sprintf("%s%s%s,%s",
		mm.prefix,
		mm.separator,
		"responseTime",
		Tags{requestMethod, requestUrl})
	mm.client.Timing(statName, responseTime)
}

func (mm *Monitor) ResponseStatus(requestMethod string, requestUrl string, responseCode int) {
	statName := fmt.Sprintf("%s%s%s,statusCode=%d,%s",
		mm.prefix,
		mm.separator,
		"responseStatusCode",
		responseCode,
		Tags{requestMethod, requestUrl})
	mm.client.Increment(statName)
}

func (mm *Monitor) Duration(operation string, d int64) {
	statName := fmt.Sprintf("%s%s%s,operation=%s", mm.prefix, mm.separator, "duration", operation)
	mm.client.Timing(statName, d)
}

func NewMonitor(client Client, prefix string, separator string) Monitor {
	return Monitor{
		client:    client,
		prefix:    prefix,
		separator: separator,
	}
}

func NewStatsdClient(statsdAddress string) *statsd.StatsdClient {
	statsdHost, statsdPortStr, _ := net.SplitHostPort(statsdAddress)
	statsdPort, _ := strconv.Atoi(statsdPortStr)
	return statsd.New(statsdHost, statsdPort)
}
