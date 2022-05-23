package metrics

import (
	"fmt"
	"net"
	"strconv"

	statsd "github.com/etsy/statsd/examples/go"
)

type StatsDConfig struct {
	Enabled bool   `mapstructure:"enabled" default:"false"`
	Address string `mapstructure:"address" default:"127.0.0.1:8125"`
	Prefix  string `mapstructure:"prefix" default:"compassApi"`
}

//go:generate mockery --name=StatsDClient -r --case underscore --with-expecter --structname StatsDClient --filename statsd_monitor.go --output=./mocks
type StatsDClient interface {
	Timing(string, int64)
	Increment(string)
}

func NewStatsDClient(statsdAddress string) *statsd.StatsdClient {
	statsdHost, statsdPortStr, _ := net.SplitHostPort(statsdAddress)
	statsdPort, _ := strconv.Atoi(statsdPortStr)
	return statsd.New(statsdHost, statsdPort)
}

type StatsDMonitor struct {
	client    StatsDClient
	prefix    string
	separator string
}

func NewStatsDMonitor(client StatsDClient, prefix string, separator string) *StatsDMonitor {
	return &StatsDMonitor{
		client:    client,
		prefix:    prefix,
		separator: separator,
	}
}

func (mm *StatsDMonitor) Duration(operation string, duration int) {
	statName := fmt.Sprintf("%s%s%s,operation=%s", mm.prefix, mm.separator, "duration", operation)
	mm.client.Timing(statName, int64(duration))
}

func (mm *StatsDMonitor) ResponseTime(requestMethod string, requestUrl string, responseTime int64) {
	statName := fmt.Sprintf("%s%s%s,%s",
		mm.prefix,
		mm.separator,
		"responseTime",
		Tags{requestMethod, requestUrl})
	mm.client.Timing(statName, responseTime)
}

func (mm *StatsDMonitor) ResponseStatus(requestMethod string, requestUrl string, responseCode int) {
	statName := fmt.Sprintf("%s%s%s,statusCode=%d,%s",
		mm.prefix,
		mm.separator,
		"responseStatusCode",
		responseCode,
		Tags{requestMethod, requestUrl})
	mm.client.Increment(statName)
}

func (mm *StatsDMonitor) ResponseTimeGRPC(FullMethod string, responseTime int64) {
	statName := fmt.Sprintf("%s%s%s,%s",
		mm.prefix,
		mm.separator,
		"responseTime",
		fmt.Sprintf("method=%s", FullMethod))
	mm.client.Timing(statName, responseTime)
}

func (mm *StatsDMonitor) ResponseStatusGRPC(fullMethod string, statusString string) {
	statName := fmt.Sprintf("%s%s%s,statusCode=%s,%s",
		mm.prefix,
		mm.separator,
		"responseStatusCode",
		statusString,
		fmt.Sprintf("method=%s", fullMethod))
	mm.client.Increment(statName)
}
