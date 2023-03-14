package statsd

import (
	"fmt"

	"github.com/goto/salt/log"
)

// Metric represents a statsd metric.
type Metric struct {
	logger        log.Logger
	name          string
	rate          float64
	tags          map[string]string
	withInfluxTag bool
	publishFunc   func(name string, tags []string, rate float64) error
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
func (m *Metric) Failure(err error) *Metric {
	if m == nil {
		return m
	}
	m.Tag("success", "false")
	return m
}

// Tag adds a tag to the metric.
func (m *Metric) Tag(key string, val string) *Metric {
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

	if m.tags == nil {
		m.tags = map[string]string{}
	}

	var ddTags []string
	if m.withInfluxTag {
		m.name = m.processTagsInflux(m.name, m.tags)
	} else {
		ddTags = m.processTagsDatadog()
	}
	go func() {
		if err := m.publishFunc(m.name, ddTags, m.rate); err != nil {
			m.logger.Warn("failed to publish metric", "name", m.name, "err", err)
		}
	}()
}

func (m *Metric) processTagsDatadog() []string {
	tags := []string{}
	for k, v := range m.tags {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}
	return tags
}

func (m *Metric) processTagsInflux(name string, tags map[string]string) string {
	var finalName = name
	for k, v := range m.tags {
		finalName = fmt.Sprintf("%s,%s=%s", finalName, k, v)
	}
	return finalName
}
