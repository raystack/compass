package asset

import "testing"

func TestTypeString(t *testing.T) {
	stringVal := TypeDashboard.String()
	if stringVal != "dashboard" {
		t.Fatalf("type dashboard converted to %s instead of 'dashboard'", stringVal)
	}
	stringVal = TypeJob.String()
	if stringVal != "job" {
		t.Fatalf("type job converted to %s instead of 'job'", stringVal)
	}
	stringVal = TypeTable.String()
	if stringVal != "table" {
		t.Fatalf("type table converted to %s instead of 'table'", stringVal)
	}
	stringVal = TypeTopic.String()
	if stringVal != "topic" {
		t.Fatalf("type topic converted to %s instead of 'topic'", stringVal)
	}
}

func TestTypeIsValid(t *testing.T) {
	aType := Type("dashboard")
	if !aType.IsValid() {
		t.Fatalf("type %s is not valid", aType)
	}
	aType = Type("job")
	if !aType.IsValid() {
		t.Fatalf("type %s is not valid", aType)
	}
	aType = Type("table")
	if !aType.IsValid() {
		t.Fatalf("type %s is not valid", aType)
	}
	aType = Type("topic")
	if !aType.IsValid() {
		t.Fatalf("type %s is not valid", aType)
	}

	aType = Type("random")
	if aType.IsValid() {
		t.Fatalf("type %s should not be valid", aType)
	}
}
