package entity

import (
	"testing"
	"time"
)

func TestType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		typ  Type
		want bool
	}{
		{"empty string is invalid", "", false},
		{"table is valid", TypeTable, true},
		{"custom type is valid", Type("satellite"), true},
		{"any non-empty string is valid", Type("anything"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.IsValid(); got != tt.want {
				t.Errorf("Type.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestType_String(t *testing.T) {
	if got := TypeTable.String(); got != "table" {
		t.Errorf("Type.String() = %q, want %q", got, "table")
	}
}

func TestEntity_IsCurrent(t *testing.T) {
	now := time.Now()

	current := Entity{ValidTo: nil}
	if !current.IsCurrent() {
		t.Error("entity with nil ValidTo should be current")
	}

	expired := Entity{ValidTo: &now}
	if expired.IsCurrent() {
		t.Error("entity with non-nil ValidTo should not be current")
	}
}
