package statsd

import (
	"fmt"

	"github.com/goto/salt/log"
)

// Metric represents a statsd metric.
type Metric struct {
	logger      log.Logger
	name        string
	rate        float64
	tags        map[string]string
	publishFunc func(name string, tags []string, rate float64) error
}

// Success tags the metric as successful.
func (m *Metric) Success() *Metric {
	if m == nil {
		return m
	}
	m.Tag("success", "true")
	return m
}

// Failure tags the metric as failure.
func (m *Metric) Failure() *Metric {
	if m == nil {
		return m
	}
	m.Tag("success", "false")
	return m
}

// Tag adds a tag to the metric.
func (m *Metric) Tag(key, val string) *Metric {
	if m == nil {
		return nil
	}

	if m.tags == nil {
		m.tags = map[string]string{}
	}

	m.tags[key] = val
	return m
}

// Publish publishes the metric with collected tags. Intended to
// be used with defer.
func (m *Metric) Publish() {
	if m == nil {
		return
	}

	tags := m.processTags()
	go func() {
		if err := m.publishFunc(m.name, tags, m.rate); err != nil {
			m.logger.Warn("failed to publish metric", "name", m.name, "err", err)
		}
	}()
}

func (m *Metric) processTags() []string {
	var tags []string
	for k, v := range m.tags {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}
	return tags
}
